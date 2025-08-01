package shared

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/emiller/tasksh/internal/estimation"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// TaskAnalyzer provides utilities for analyzing and categorizing tasks
type TaskAnalyzer struct {
	timeDB    *timedb.TimeDB
	estimator *estimation.Estimator
}

// NewTaskAnalyzer creates a new task analyzer
func NewTaskAnalyzer(timeDB *timedb.TimeDB) *TaskAnalyzer {
	analyzer := &TaskAnalyzer{timeDB: timeDB}
	
	// Create estimator if we have a timeDB
	if timeDB != nil {
		config := estimation.DefaultConfig()
		config.PreferTimewarrior = true
		analyzer.estimator = estimation.NewEstimator(timeDB, config)
	}
	
	return analyzer
}

// AnalyzeTask converts a taskwarrior task into a planned task with metadata
func (ta *TaskAnalyzer) AnalyzeTask(task *taskwarrior.Task, targetDate time.Time) PlannedTask {
	planned := PlannedTask{
		Task:        task,
		PlannedDate: targetDate,
	}

	// Calculate urgency
	planned.Urgency = ta.calculateUrgency(task)

	// Determine category based on urgency and due date
	planned.Category = ta.categorizeTask(planned)

	// Determine scheduling status
	planned.IsScheduled = task.Due != ""
	planned.IsDue = ta.isDueToday(task, targetDate)

	// Estimate time required
	planned.EstimatedHours, planned.EstimationReason = ta.estimateTaskTime(task)

	// Determine optimal time slot and required energy
	planned.RequiredEnergy = ta.calculateRequiredEnergy(task)
	planned.OptimalTimeSlot = ta.getOptimalTimeSlot(planned)

	return planned
}

// calculateUrgency calculates a task's urgency score
func (ta *TaskAnalyzer) calculateUrgency(task *taskwarrior.Task) float64 {
	urgency := 0.0

	// Priority contribution
	switch task.Priority {
	case "H":
		urgency += 6.0
	case "M":
		urgency += 1.8
	case "L":
		urgency += 0.0
	}

	// Due date contribution
	if task.Due != "" {
		if dueTime, err := time.Parse("2006-01-02T15:04:05Z", task.Due); err == nil {
			daysUntilDue := time.Until(dueTime).Hours() / 24
			if daysUntilDue <= 0 {
				urgency += 15.0 // Overdue
			} else if daysUntilDue <= 1 {
				urgency += 12.0 // Due today/tomorrow
			} else if daysUntilDue <= 7 {
				urgency += 6.0 // Due this week
			} else if daysUntilDue <= 30 {
				urgency += 2.0 // Due this month
			}
		}
	}

	// Project contribution
	if task.Project != "" {
		urgency += 1.0
	}

	return urgency
}

// categorizeTask determines the category based on urgency and due date
func (ta *TaskAnalyzer) categorizeTask(task PlannedTask) TaskCategory {
	// Critical: Due today/overdue or very high urgency
	if task.IsDue || task.Urgency >= 20.0 {
		return CategoryCritical
	}

	// Important: Moderate to high urgency
	if task.Urgency >= 10.0 {
		return CategoryImportant
	}

	// Everything else is flexible
	return CategoryFlexible
}

// isDueToday checks if a task is due on the target date
func (ta *TaskAnalyzer) isDueToday(task *taskwarrior.Task, targetDate time.Time) bool {
	if task.Due == "" {
		return false
	}

	dueTime, err := time.Parse("2006-01-02T15:04:05Z", task.Due)
	if err != nil {
		return false
	}

	targetStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	targetEnd := targetStart.Add(24 * time.Hour)

	return dueTime.After(targetStart) && dueTime.Before(targetEnd)
}

// estimateTaskTime estimates the time required for a task
func (ta *TaskAnalyzer) estimateTaskTime(task *taskwarrior.Task) (float64, string) {
	// Use the new multi-source estimator if available
	if ta.estimator != nil {
		estimate, err := ta.estimator.EstimateTask(task)
		if err == nil && estimate != nil {
			// Add small buffer for tasks over 30 minutes
			bufferedHours := estimate.Hours
			bufferText := ""
			if estimate.Hours > 0.5 {
				bufferedHours = estimate.Hours * 1.1 // 10% buffer for longer tasks
				bufferText = " (with 10% buffer)"
			}
			confidence := fmt.Sprintf(" (%.0f%% confidence)", estimate.Confidence*100)
			return bufferedHours, estimate.Reason + confidence + bufferText
		}
	}
	
	// Fallback to old method
	if ta.timeDB == nil {
		return ta.fallbackEstimate(task), "Default estimate (no historical data)"
	}

	estimate, reason, err := ta.timeDB.EstimateTimeForTask(task)
	if err != nil || estimate <= 0 {
		fallbackHours := ta.fallbackEstimate(task)
		return fallbackHours, ta.getEstimateReason(task, fallbackHours)
	}

	// Add small buffer for longer estimates
	if estimate > 0.5 {
		bufferedEstimate := estimate * 1.1
		return bufferedEstimate, reason + " (with 10% buffer)"
	}
	return estimate, reason
}

// fallbackEstimate provides fallback time estimates based on task characteristics
func (ta *TaskAnalyzer) fallbackEstimate(task *taskwarrior.Task) float64 {
	description := strings.ToLower(task.Description)
	
	// Very quick tasks (5-15 minutes = 0.08-0.25 hours)
	veryQuickKeywords := []string{
		"brush teeth", "take medication", "take heart medication", "drink water", "stretch",
		"check weather", "lock door", "feed pet", "water plants",
		"take vitamins", "quick break", "bathroom", "get coffee",
	}
	for _, keyword := range veryQuickKeywords {
		if strings.Contains(description, keyword) {
			return 0.1 // 6 minutes
		}
	}
	
	// Quick tasks (15-30 minutes = 0.25-0.5 hours)
	quickKeywords := []string{
		"email", "quick call", "respond", "reply", "check", "update status",
		"fix typo", "rename", "quick fix", "stand up", "daily standup",
		"check mail", "take out trash", "dishes", "make bed",
	}
	for _, keyword := range quickKeywords {
		if strings.Contains(description, keyword) {
			return 0.25 // 15 minutes
		}
	}
	
	// Long tasks (2-4 hours) - check first for more specific matches
	longKeywords := []string{
		"research", "analyze", "investigate", "plan project", "plan workshop", "architecture",
		"integrate", "migration", "presentation", "training",
	}
	for _, keyword := range longKeywords {
		if strings.Contains(description, keyword) {
			return 2.5 // 2.5 hours
		}
	}
	
	// Medium tasks (1-2 hours)
	mediumKeywords := []string{
		"implement feature", "create", "design", "develop", "build",
		"write proposal", "major bug", "refactor", "meeting", "interview",
		"deep work", "focus time", "study", "learn", "practice",
		"comprehensive documentation", "create documentation", "workshop",
	}
	for _, keyword := range mediumKeywords {
		if strings.Contains(description, keyword) {
			return 1.5 // 1.5 hours
		}
	}
	
	// Short tasks (30-60 minutes = 0.5-1 hour)
	shortKeywords := []string{
		"review pr", "test", "document", "clean", "organize", "minor",
		"small bug", "update docs", "write notes", "prep", "prepare",
		"grocery", "errands", "workout", "exercise", "walk", "cook",
	}
	for _, keyword := range shortKeywords {
		if strings.Contains(description, keyword) {
			return 0.5 // 30 minutes
		}
	}
	
	// Check for specific patterns
	if strings.Contains(description, "call") && !strings.Contains(description, "quick") {
		return 0.5 // Regular calls are 30 min
	}
	
	// Default estimate based on description length as last resort
	// (Priority doesn't determine duration - urgent != long)
	if len(description) < 15 {
		return 0.25 // Very short = quick task
	} else if len(description) < 30 {
		return 0.5 // Short description = 30 min
	} else if len(description) < 60 {
		return 1.0 // Medium description = 1 hour
	} else {
		return 1.5 // Long description = 1.5 hours
	}
}

// getEstimateReason provides a human-readable reason for the time estimate
func (ta *TaskAnalyzer) getEstimateReason(task *taskwarrior.Task, hours float64) string {
	description := strings.ToLower(task.Description)
	
	// Check for very quick tasks
	veryQuickKeywords := []string{"brush teeth", "take medication", "drink water", "stretch", "feed pet"}
	for _, keyword := range veryQuickKeywords {
		if strings.Contains(description, keyword) {
			return "Very quick personal task"
		}
	}
	
	// Check for keyword matches
	if strings.Contains(description, "email") || strings.Contains(description, "quick call") {
		return "Quick communication task"
	}
	if strings.Contains(description, "review") || strings.Contains(description, "check") {
		return "Review/verification task"
	}
	if strings.Contains(description, "implement") || strings.Contains(description, "create") || strings.Contains(description, "build") {
		return "Development/creation task"
	}
	if strings.Contains(description, "research") || strings.Contains(description, "analyze") {
		return "Research/analysis task"
	}
	if strings.Contains(description, "meeting") || strings.Contains(description, "discuss") {
		return "Meeting/collaboration task"
	}
	if strings.Contains(description, "workout") || strings.Contains(description, "exercise") {
		return "Physical activity"
	}
	if strings.Contains(description, "errands") || strings.Contains(description, "grocery") {
		return "Errand/shopping task"
	}
	
	// Time-based reasons
	if hours <= 0.1 {
		return "Very quick routine task"
	} else if hours <= 0.25 {
		return "Quick task (15 min)"
	} else if hours <= 0.5 {
		return "Short task (30 min)"
	} else if hours <= 1.0 {
		return "Standard task (1 hour)"
	} else if hours <= 2.0 {
		return "Extended task (1-2 hours)"
	} else {
		return "Long task (2+ hours)"
	}
}

// calculateRequiredEnergy determines the energy level required for a task
func (ta *TaskAnalyzer) calculateRequiredEnergy(task *taskwarrior.Task) EnergyLevel {
	// High energy indicators
	highEnergyKeywords := []string{"design", "code", "develop", "write", "create", "analyze", "research", "plan", "architect", "review", "think", "strategy"}
	
	// Low energy indicators  
	lowEnergyKeywords := []string{"email", "call", "meeting", "standup", "sync", "update", "check", "status", "admin", "file", "schedule"}
	
	description := strings.ToLower(task.Description)
	
	// Check for high energy keywords
	for _, keyword := range highEnergyKeywords {
		if strings.Contains(description, keyword) {
			return EnergyHigh
		}
	}
	
	// Check for low energy keywords
	for _, keyword := range lowEnergyKeywords {
		if strings.Contains(description, keyword) {
			return EnergyLow
		}
	}
	
	// Consider task priority and estimated duration
	if task.Priority == "H" {
		return EnergyHigh
	}
	
	// Default to medium energy
	return EnergyMedium
}

// getOptimalTimeSlot suggests the best time of day for a task
func (ta *TaskAnalyzer) getOptimalTimeSlot(task PlannedTask) string {
	switch task.RequiredEnergy {
	case EnergyHigh:
		return "morning" // High focus work best in morning
	case EnergyMedium:
		if task.Category == CategoryCritical {
			return "morning"
		}
		return "afternoon"
	case EnergyLow:
		return "anytime" // Low energy tasks can be done anytime
	default:
		return "afternoon"
	}
}

// LoadTasksByFilter loads tasks using taskwarrior filters
func LoadTasksByFilter(filters ...string) ([]string, error) {
	args := append([]string{
		"rc.color=off",
		"rc.detection=off",
		"rc._forcecolor=off",
		"rc.verbose=nothing",
	}, filters...)
	args = append(args, "uuids")

	cmd := exec.Command("task", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("task command failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	return strings.Fields(result), nil
}

// SortTasksByPriority sorts planned tasks by planning priority
func SortTasksByPriority(tasks []PlannedTask) {
	sort.Slice(tasks, func(i, j int) bool {
		taskA, taskB := tasks[i], tasks[j]

		// Critical tasks first
		if taskA.Category != taskB.Category {
			return taskA.Category < taskB.Category // Critical = 0, Important = 1, Flexible = 2
		}

		// Within same category, due tasks first
		if taskA.IsDue && !taskB.IsDue {
			return true
		}
		if !taskA.IsDue && taskB.IsDue {
			return false
		}

		// Then by urgency (higher first)
		if taskA.Urgency != taskB.Urgency {
			return taskA.Urgency > taskB.Urgency
		}

		// Finally by description for stable sort
		return taskA.Description < taskB.Description
	})
}

// CalculateWorkloadSummary calculates workload statistics for a set of tasks
func CalculateWorkloadSummary(tasks []PlannedTask) (totalHours float64, breakdown map[TaskCategory]float64, energyBreakdown map[EnergyLevel]float64) {
	breakdown = make(map[TaskCategory]float64)
	energyBreakdown = make(map[EnergyLevel]float64)

	for _, task := range tasks {
		totalHours += task.EstimatedHours
		breakdown[task.Category] += task.EstimatedHours
		energyBreakdown[task.RequiredEnergy] += task.EstimatedHours
	}

	return totalHours, breakdown, energyBreakdown
}