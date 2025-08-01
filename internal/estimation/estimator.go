package estimation

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	
	"github.com/emiller/tasksh/internal/ai"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
	"github.com/emiller/tasksh/internal/timewarrior"
)

// EstimationSource represents where an estimate came from
type EstimationSource string

const (
	SourceTimewarrior EstimationSource = "timewarrior"
	SourceHistorical  EstimationSource = "historical"
	SourceAI          EstimationSource = "ai"
	SourceKeywords    EstimationSource = "keywords"
	SourceDefault     EstimationSource = "default"
)

// Estimate represents a time estimate with metadata
type Estimate struct {
	Hours      float64
	Source     EstimationSource
	Confidence float64 // 0.0 to 1.0
	Reason     string
	Details    map[string]interface{}
}

// Estimator combines multiple sources for time estimation
type Estimator struct {
	timeDB      *timedb.TimeDB
	timeWarrior *timewarrior.Client
	aiAnalyzer  *ai.Analyzer
	aiCache     *AICache
	config      *EstimatorConfig
	lastSync    time.Time
	syncMutex   sync.Mutex
}

// EstimatorConfig configures the estimator behavior
type EstimatorConfig struct {
	UseAI              bool
	AIEnabled          bool
	PreferTimewarrior  bool
	MinConfidence      float64
	CacheAIEstimates   bool
	FallbackHours      float64
	AutoSyncInterval   time.Duration // How often to auto-sync timewarrior data
	AutoSyncEnabled    bool          // Whether to automatically sync in background
}

// DefaultConfig returns default estimator configuration
func DefaultConfig() *EstimatorConfig {
	return &EstimatorConfig{
		UseAI:             false,
		AIEnabled:         false,
		PreferTimewarrior: true,
		MinConfidence:     0.3,
		CacheAIEstimates:  true,
		FallbackHours:     0.5, // 30 minutes is a reasonable default
		AutoSyncInterval:  4 * time.Hour, // Sync every 4 hours
		AutoSyncEnabled:   true, // Enable by default
	}
}

// NewEstimator creates a new multi-source estimator
func NewEstimator(timeDB *timedb.TimeDB, config *EstimatorConfig) *Estimator {
	if config == nil {
		config = DefaultConfig()
	}
	
	estimator := &Estimator{
		timeDB:      timeDB,
		timeWarrior: timewarrior.NewClient(),
		config:      config,
		lastSync:    time.Time{}, // Zero time means never synced
	}
	
	if config.AIEnabled {
		estimator.aiAnalyzer = ai.NewAnalyzer(timeDB)
		
		// Initialize AI cache if caching is enabled
		if config.CacheAIEstimates {
			// Use same directory as timeDB
			homeDir, _ := os.UserHomeDir()
			cacheDir := filepath.Join(homeDir, ".local", "share", "tasksh")
			cachePath := filepath.Join(cacheDir, "ai_cache.sqlite3")
			
			if cache, err := NewAICache(cachePath); err == nil {
				estimator.aiCache = cache
			}
		}
	}
	
	// Check last sync time from database
	if timeDB != nil {
		if lastSync, err := timeDB.GetLastSyncTime(); err == nil && !lastSync.IsZero() {
			estimator.lastSync = lastSync
		}
	}
	
	return estimator
}

// EstimateTask provides the best time estimate for a task
func (e *Estimator) EstimateTask(task *taskwarrior.Task) (*Estimate, error) {
	// Auto-sync if enabled and needed
	if e.config.AutoSyncEnabled {
		e.autoSyncIfNeeded()
	}
	
	// Collect estimates from all sources
	estimates := e.collectEstimates(task)
	
	// Sort by confidence and source preference
	sort.Slice(estimates, func(i, j int) bool {
		// Prefer timewarrior if configured
		if e.config.PreferTimewarrior {
			if estimates[i].Source == SourceTimewarrior && estimates[j].Source != SourceTimewarrior {
				return true
			}
			if estimates[i].Source != SourceTimewarrior && estimates[j].Source == SourceTimewarrior {
				return false
			}
		}
		
		// Otherwise sort by confidence
		return estimates[i].Confidence > estimates[j].Confidence
	})
	
	// Find the best estimate above minimum confidence
	var bestEstimate *Estimate
	for _, estimate := range estimates {
		if estimate.Confidence >= e.config.MinConfidence {
			bestEstimate = &estimate
			break
		}
	}
	
	// Use fallback if no good estimates
	if bestEstimate == nil {
		bestEstimate = &Estimate{
			Hours:      e.config.FallbackHours,
			Source:     SourceDefault,
			Confidence: 0.1,
			Reason:     "No reliable estimates available, using default",
		}
	}
	
	// Apply learning-based adjustments
	bestEstimate = e.ApplyLearning(bestEstimate, task)
	
	return bestEstimate, nil
}

// collectEstimates gathers estimates from all available sources
func (e *Estimator) collectEstimates(task *taskwarrior.Task) []Estimate {
	estimates := []Estimate{}
	
	// 1. Timewarrior actual time (if task has been worked on)
	if twEstimate := e.getTimewarriorEstimate(task); twEstimate != nil {
		estimates = append(estimates, *twEstimate)
	}
	
	// 2. Historical database with enhanced similarity
	if dbEstimate := e.getDatabaseEstimate(task); dbEstimate != nil {
		estimates = append(estimates, *dbEstimate)
	}
	
	// 3. AI-powered estimation (if enabled)
	if e.config.AIEnabled && e.aiAnalyzer != nil {
		if aiEstimate := e.getAIEstimate(task); aiEstimate != nil {
			estimates = append(estimates, *aiEstimate)
		}
	}
	
	// 4. Keyword-based fallback
	if kwEstimate := e.getKeywordEstimate(task); kwEstimate != nil {
		estimates = append(estimates, *kwEstimate)
	}
	
	return estimates
}

// getTimewarriorEstimate checks if task has existing timewarrior data
func (e *Estimator) getTimewarriorEstimate(task *taskwarrior.Task) *Estimate {
	entries, err := e.timeWarrior.GetEntriesForTask(task.UUID)
	if err != nil || len(entries) == 0 {
		return nil
	}
	
	// Calculate total hours tracked
	totalHours := timewarrior.CalculateTotalHours(entries)
	if totalHours == 0 {
		return nil
	}
	
	// Check if task is complete or still in progress
	isComplete := task.Status == "completed"
	confidence := 0.9
	reason := fmt.Sprintf("Already tracked %.1f hours", totalHours)
	
	if !isComplete {
		// Task still in progress, estimate additional time needed
		// Use historical data to estimate completion percentage
		avgCompletion := e.estimateCompletionPercentage(task, totalHours)
		if avgCompletion > 0 {
			estimatedTotal := totalHours / avgCompletion
			reason = fmt.Sprintf("Tracked %.1f hours (estimated %.0f%% complete)", 
				totalHours, avgCompletion*100)
			totalHours = estimatedTotal
			confidence = 0.7
		} else {
			reason += " (in progress)"
			confidence = 0.5
		}
	}
	
	return &Estimate{
		Hours:      totalHours,
		Source:     SourceTimewarrior,
		Confidence: confidence,
		Reason:     reason,
		Details: map[string]interface{}{
			"entries":     len(entries),
			"is_complete": isComplete,
		},
	}
}

// getDatabaseEstimate uses enhanced similarity matching
func (e *Estimator) getDatabaseEstimate(task *taskwarrior.Task) *Estimate {
	hours, reason, confidence, err := e.timeDB.EstimateBasedOnSimilarity(task)
	if err != nil || hours == 0 {
		return nil
	}
	
	return &Estimate{
		Hours:      hours,
		Source:     SourceHistorical,
		Confidence: confidence,
		Reason:     reason,
	}
}

// getAIEstimate uses AI for estimation (with caching)
func (e *Estimator) getAIEstimate(task *taskwarrior.Task) *Estimate {
	// Check cache first if enabled
	if e.config.CacheAIEstimates && e.aiCache != nil {
		cached, err := e.aiCache.Get(task.Description, task.Project)
		if err == nil && cached != nil {
			return cached
		}
	}
	
	// Get AI analysis
	analysis, err := e.aiAnalyzer.AnalyzeTask(task)
	if err != nil || analysis.TimeEstimate.Hours == 0 {
		return nil
	}
	
	estimate := &Estimate{
		Hours:      analysis.TimeEstimate.Hours,
		Source:     SourceAI,
		Confidence: 0.6, // Base AI confidence
		Reason:     analysis.TimeEstimate.Reason,
		Details: map[string]interface{}{
			"suggestions": len(analysis.Suggestions),
		},
	}
	
	// Cache the estimate
	if e.config.CacheAIEstimates && e.aiCache != nil {
		// Cache for 7 days
		_ = e.aiCache.Set(task.Description, task.Project, estimate, 7*24*time.Hour)
	}
	
	return estimate
}

// getKeywordEstimate provides keyword-based fallback estimation
func (e *Estimator) getKeywordEstimate(task *taskwarrior.Task) *Estimate {
	description := strings.ToLower(task.Description)
	hours := 0.0
	confidence := 0.4
	reason := ""
	
	// Very quick tasks (5-15 minutes)
	veryQuickKeywords := []string{
		"brush teeth", "take medication", "take heart medication", "drink water", "stretch",
		"check weather", "lock door", "feed pet", "water plants",
		"take vitamins", "quick break", "bathroom", "get coffee",
	}
	for _, keyword := range veryQuickKeywords {
		if strings.Contains(description, keyword) {
			hours = 0.1 // 6 minutes
			reason = "Very quick personal task"
			confidence = 0.7 // High confidence for these routine tasks
			break
		}
	}
	
	// Quick tasks (15-30 minutes)
	if hours == 0 {
		quickKeywords := []string{
			"email", "quick call", "respond", "reply", "check", "update status",
			"fix typo", "rename", "quick fix", "stand up", "daily standup",
			"check mail", "take out trash", "dishes", "make bed",
		}
		for _, keyword := range quickKeywords {
			if strings.Contains(description, keyword) {
				hours = 0.25 // 15 minutes
				reason = "Quick task based on keywords"
				confidence = 0.6
				break
			}
		}
	}
	
	// Long tasks (2-4 hours) - check first for more specific matches
	if hours == 0 {
		longKeywords := []string{
			"research", "analyze", "investigate", "plan project", "plan workshop", "architecture",
			"integrate", "migration", "presentation", "training",
		}
		for _, keyword := range longKeywords {
			if strings.Contains(description, keyword) {
				hours = 2.5
				reason = "Long task based on keywords"
				confidence = 0.4
				break
			}
		}
	}
	
	// Medium tasks (1-2 hours)
	if hours == 0 {
		mediumKeywords := []string{
			"implement feature", "create", "design", "develop", "build",
			"write proposal", "major bug", "refactor", "meeting", "interview",
			"deep work", "focus time", "study", "learn", "practice",
			"comprehensive documentation", "create documentation", "workshop",
		}
		for _, keyword := range mediumKeywords {
			if strings.Contains(description, keyword) {
				hours = 1.5
				reason = "Medium task based on keywords"
				confidence = 0.45
				break
			}
		}
	}
	
	// Short tasks (30-60 minutes)
	if hours == 0 {
		shortKeywords := []string{
			"review pr", "test", "document", "clean", "organize", "minor",
			"small bug", "update docs", "write notes", "prep", "prepare",
			"grocery", "errands", "workout", "exercise", "walk", "cook",
		}
		for _, keyword := range shortKeywords {
			if strings.Contains(description, keyword) {
				hours = 0.5 // 30 minutes
				reason = "Short task based on keywords"
				confidence = 0.5
				break
			}
		}
	}
	
	// Check for specific patterns
	if hours == 0 && strings.Contains(description, "call") && !strings.Contains(description, "quick") {
		hours = 0.5 // Regular calls are 30 min
		reason = "Standard call duration"
		confidence = 0.5
	}
	
	// Default based on description length
	// Note: Priority (H/M/L) is about urgency/importance, NOT duration!
	if hours == 0 {
		if len(description) < 15 {
			hours = 0.25
			reason = "Very short description suggests quick task"
		} else if len(description) < 30 {
			hours = 0.5
			reason = "Short description suggests simple task"
		} else if len(description) < 60 {
			hours = 1.0
			reason = "Medium description suggests standard task"
		} else {
			hours = 1.5
			reason = "Long description suggests complex task"
		}
		confidence = 0.2
	}
	
	return &Estimate{
		Hours:      hours,
		Source:     SourceKeywords,
		Confidence: confidence,
		Reason:     reason,
	}
}

// estimateCompletionPercentage estimates how complete a task is based on time tracked
func (e *Estimator) estimateCompletionPercentage(task *taskwarrior.Task, hoursTracked float64) float64 {
	// Get similar completed tasks
	similar, err := e.timeDB.GetSimilarTasksEnhanced(task, 5)
	if err != nil || len(similar) == 0 {
		return 0
	}
	
	// Calculate average completion percentage based on similar tasks
	totalPercentage := 0.0
	count := 0
	
	for _, match := range similar {
		if match.Entry.ActualHours > 0 {
			// Estimate this task is at hoursTracked/actualHours percentage
			percentage := math.Min(hoursTracked/match.Entry.ActualHours, 1.0)
			totalPercentage += percentage * match.Score // Weight by similarity
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return totalPercentage / float64(count)
}

// autoSyncIfNeeded performs background sync if enough time has passed
func (e *Estimator) autoSyncIfNeeded() {
	// Check if we need to sync (without holding the lock)
	e.syncMutex.Lock()
	needSync := time.Since(e.lastSync) > e.config.AutoSyncInterval
	e.syncMutex.Unlock()
	
	if !needSync {
		return
	}
	
	// Perform sync in background goroutine to avoid blocking
	go func() {
		// Try to acquire lock, but don't block if another sync is in progress
		if !e.syncMutex.TryLock() {
			return
		}
		defer e.syncMutex.Unlock()
		
		// Double-check we still need to sync
		if time.Since(e.lastSync) <= e.config.AutoSyncInterval {
			return
		}
		
		// Sync recent data (last 7 days for efficiency)
		since := time.Now().AddDate(0, 0, -7)
		if _, err := e.timeDB.SyncFromTimewarrior(e.timeWarrior, since); err == nil {
			e.lastSync = time.Now()
			// The database will update its own sync timestamp
		}
		// Silently ignore errors - this is background sync
	}()
}

// GetLastSyncTime returns when the data was last synced
func (e *Estimator) GetLastSyncTime() time.Time {
	e.syncMutex.Lock()
	defer e.syncMutex.Unlock()
	return e.lastSync
}

// ForceSync forces an immediate sync of timewarrior data
func (e *Estimator) ForceSync(since time.Time) error {
	e.syncMutex.Lock()
	defer e.syncMutex.Unlock()
	
	result, err := e.timeDB.SyncFromTimewarrior(e.timeWarrior, since)
	if err != nil {
		return err
	}
	
	e.lastSync = time.Now()
	fmt.Printf("Sync completed: %d new entries, %d updated\n", result.NewEntries, result.UpdatedEntries)
	return nil
}
