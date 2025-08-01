package weekly

import (
	"fmt"
	"strings"
	"time"

	"github.com/emiller/tasksh/internal/planning/shared"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// WorkflowStep represents a step in the weekly planning workflow
type WorkflowStep int

const (
	StepWeeklyReflection WorkflowStep = iota
	StepObjectiveSetting
	StepJournaling
	StepWorkStreamPlanning
	StepWeeklySummary
	StepCompleted
)

// String returns a human-readable step name
func (s WorkflowStep) String() string {
	switch s {
	case StepWeeklyReflection:
		return "Reflect on Last Week"
	case StepObjectiveSetting:
		return "Set Weekly Objectives"
	case StepJournaling:
		return "Strategic Journaling"
	case StepWorkStreamPlanning:
		return "Plan Work Streams"
	case StepWeeklySummary:
		return "Review Weekly Plan"
	case StepCompleted:
		return "Planning Complete"
	default:
		return "Unknown Step"
	}
}

// WeeklyPlanningSession represents a strategic weekly planning session
type WeeklyPlanningSession struct {
	Context         shared.PlanningContext
	CurrentStep     WorkflowStep
	WeeklyReflection *WeeklyReflectionData
	Objectives      []shared.Objective
	JournalEntry    string
	WorkStreams     []WorkStream
	AvailableTasks  []shared.PlannedTask
	
	// Session metadata
	StartedAt   time.Time
	CompletedAt *time.Time
	WeekStart   time.Time
	WeekEnd     time.Time
	
	// Internal utilities
	analyzer *shared.TaskAnalyzer
}

// WeeklyReflectionData represents reflection data for the previous week
type WeeklyReflectionData struct {
	PreviousWeekStart time.Time
	PreviousWeekEnd   time.Time
	KeyAccomplishments []string
	CompletedObjectives []string
	UnfinishedObjectives []string
	LessonsLearned    []string
	Challenges        []string
	EnergyPattern     string
	OverallSatisfaction int // 1-5 scale
	FocusAreas        []string
}

// WorkStream represents a thematic area of work for the week
type WorkStream struct {
	ID          string
	Title       string
	Description string
	ObjectiveIDs []string // Links to objectives this stream supports
	EstimatedHours float64
	Priority    int
	Tasks       []shared.PlannedTask
	TimeBlocks  []TimeBlock
}

// TimeBlock represents a planned time allocation
type TimeBlock struct {
	Day         time.Weekday
	StartHour   int
	EndHour     int
	Description string
}

// NewWeeklyPlanningSession creates a new weekly planning session
func NewWeeklyPlanningSession(weekStart time.Time) (*WeeklyPlanningSession, error) {
	timeDB, err := timedb.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open time database: %w", err)
	}

	// Calculate week boundaries
	weekEnd := weekStart.AddDate(0, 0, 6)

	context := shared.PlanningContext{
		Date:         weekStart,
		TimeDB:       timeDB,
		WorkdayHours: 8.0,
		EnergyLevel:  shared.EnergyMedium,
	}

	session := &WeeklyPlanningSession{
		Context:     context,
		CurrentStep: StepWeeklyReflection,
		StartedAt:   time.Now(),
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
		analyzer:    shared.NewTaskAnalyzer(timeDB),
	}

	return session, nil
}

// Close closes the planning session and releases resources
func (s *WeeklyPlanningSession) Close() error {
	if s.Context.TimeDB != nil {
		return s.Context.TimeDB.Close()
	}
	return nil
}

// GetStepProgress returns the current step and total steps
func (s *WeeklyPlanningSession) GetStepProgress() (current int, total int) {
	return int(s.CurrentStep), int(StepCompleted)
}

// IsCompleted returns true if the planning session is completed
func (s *WeeklyPlanningSession) IsCompleted() bool {
	return s.CurrentStep == StepCompleted
}

// NextStep advances to the next step in the workflow
func (s *WeeklyPlanningSession) NextStep() error {
	switch s.CurrentStep {
	case StepWeeklyReflection:
		s.CurrentStep = StepObjectiveSetting
	case StepObjectiveSetting:
		s.CurrentStep = StepJournaling
	case StepJournaling:
		if err := s.loadRelevantTasks(); err != nil {
			return fmt.Errorf("failed to load tasks: %w", err)
		}
		s.CurrentStep = StepWorkStreamPlanning
	case StepWorkStreamPlanning:
		s.CurrentStep = StepWeeklySummary
	case StepWeeklySummary:
		s.CurrentStep = StepCompleted
		completedAt := time.Now()
		s.CompletedAt = &completedAt
	default:
		return fmt.Errorf("cannot advance from current step: %s", s.CurrentStep)
	}
	return nil
}

// PreviousStep goes back to the previous step (if allowed)
func (s *WeeklyPlanningSession) PreviousStep() error {
	switch s.CurrentStep {
	case StepObjectiveSetting:
		s.CurrentStep = StepWeeklyReflection
	case StepJournaling:
		s.CurrentStep = StepObjectiveSetting
	case StepWorkStreamPlanning:
		s.CurrentStep = StepJournaling
	case StepWeeklySummary:
		s.CurrentStep = StepWorkStreamPlanning
	default:
		return fmt.Errorf("cannot go back from current step: %s", s.CurrentStep)
	}
	return nil
}

// SetWeeklyReflection sets the reflection data for the previous week
func (s *WeeklyPlanningSession) SetWeeklyReflection(reflection *WeeklyReflectionData) {
	s.WeeklyReflection = reflection
}

// GetReflectionPrompts returns prompts for weekly reflection
func (s *WeeklyPlanningSession) GetReflectionPrompts() []string {
	lastWeekStart := s.WeekStart.AddDate(0, 0, -7)
	lastWeekEnd := s.WeekStart.AddDate(0, 0, -1)
	
	return []string{
		fmt.Sprintf("What were your key accomplishments last week (%s - %s)?", 
			lastWeekStart.Format("Jan 2"), lastWeekEnd.Format("Jan 2")),
		"Which objectives or goals did you complete?",
		"What remained unfinished and why?",
		"What lessons did you learn about your work or productivity?",
		"What challenges did you face?",
		"How was your energy and focus throughout the week?",
		"What would you do differently?",
	}
}

// AddObjective adds a new objective for the week
func (s *WeeklyPlanningSession) AddObjective(title, description string, priority int) *shared.Objective {
	objective := shared.Objective{
		ID:           fmt.Sprintf("obj_%d", time.Now().Unix()),
		Title:        title,
		Description:  description,
		Priority:     priority,
		CreatedAt:    time.Now(),
		Status:       shared.ObjectiveActive,
		EstimatedWeeks: 1, // Default for weekly objective
	}
	
	s.Objectives = append(s.Objectives, objective)
	return &objective
}

// RemoveObjective removes an objective by index
func (s *WeeklyPlanningSession) RemoveObjective(index int) error {
	if index < 0 || index >= len(s.Objectives) {
		return fmt.Errorf("invalid objective index: %d", index)
	}
	
	s.Objectives = append(s.Objectives[:index], s.Objectives[index+1:]...)
	return nil
}

// UpdateObjective updates an existing objective
func (s *WeeklyPlanningSession) UpdateObjective(index int, title, description string) error {
	if index < 0 || index >= len(s.Objectives) {
		return fmt.Errorf("invalid objective index: %d", index)
	}
	
	s.Objectives[index].Title = title
	s.Objectives[index].Description = description
	return nil
}

// SetJournalEntry sets the strategic journal entry
func (s *WeeklyPlanningSession) SetJournalEntry(entry string) {
	s.JournalEntry = strings.TrimSpace(entry)
}

// loadRelevantTasks loads tasks relevant for the week
func (s *WeeklyPlanningSession) loadRelevantTasks() error {
	weekEndStr := s.WeekEnd.Format("2006-01-02")
	
	// Get tasks due this week or with moderate urgency
	filters := []string{
		fmt.Sprintf("(due.before:%s or urgency.over:8.0)", weekEndStr),
		"and", "(+PENDING", "or", "+WAITING)",
	}

	uuids, err := shared.LoadTasksByFilter(filters...)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Batch load all tasks
	taskMap, err := taskwarrior.BatchLoadTasks(uuids)
	if err != nil {
		return fmt.Errorf("failed to batch load tasks: %w", err)
	}

	// Convert to PlannedTasks with analysis
	s.AvailableTasks = make([]shared.PlannedTask, 0, len(uuids))
	for _, uuid := range uuids {
		task, ok := taskMap[uuid]
		if !ok {
			continue
		}

		plannedTask := s.analyzer.AnalyzeTask(task, s.WeekStart)
		s.AvailableTasks = append(s.AvailableTasks, plannedTask)
	}

	// Sort by priority
	shared.SortTasksByPriority(s.AvailableTasks)

	return nil
}

// CreateWorkStream creates a new work stream
func (s *WeeklyPlanningSession) CreateWorkStream(title, description string, objectiveIDs []string, estimatedHours float64) *WorkStream {
	workStream := WorkStream{
		ID:             fmt.Sprintf("ws_%d", time.Now().Unix()),
		Title:          title,
		Description:    description,
		ObjectiveIDs:   objectiveIDs,
		EstimatedHours: estimatedHours,
		Priority:       len(s.WorkStreams) + 1,
		Tasks:          []shared.PlannedTask{},
		TimeBlocks:     []TimeBlock{},
	}
	
	s.WorkStreams = append(s.WorkStreams, workStream)
	return &workStream
}

// AddTaskToWorkStream adds a task to a work stream
func (s *WeeklyPlanningSession) AddTaskToWorkStream(workStreamIndex, taskIndex int) error {
	if workStreamIndex < 0 || workStreamIndex >= len(s.WorkStreams) {
		return fmt.Errorf("invalid work stream index: %d", workStreamIndex)
	}
	
	if taskIndex < 0 || taskIndex >= len(s.AvailableTasks) {
		return fmt.Errorf("invalid task index: %d", taskIndex)
	}
	
	task := s.AvailableTasks[taskIndex]
	s.WorkStreams[workStreamIndex].Tasks = append(s.WorkStreams[workStreamIndex].Tasks, task)
	
	return nil
}

// GetObjectiveSuggestions returns AI-powered suggestions for weekly objectives
func (s *WeeklyPlanningSession) GetObjectiveSuggestions() []string {
	// Simple heuristic-based suggestions (could be enhanced with AI)
	suggestions := []string{
		"Complete high-priority project milestone",
		"Improve team communication and collaboration",
		"Focus on skill development and learning",
		"Optimize workflow and productivity systems",
		"Strengthen strategic partnerships",
		"Advance key business objectives",
	}
	
	// Add task-based suggestions if we have loaded tasks
	if len(s.AvailableTasks) > 0 {
		projectMap := make(map[string]int)
		for _, task := range s.AvailableTasks {
			if task.Project != "" {
				projectMap[task.Project]++
			}
		}
		
		// Suggest objectives based on projects with multiple tasks
		for project, count := range projectMap {
			if count >= 3 {
				suggestions = append(suggestions, fmt.Sprintf("Make significant progress on %s project", project))
			}
		}
	}
	
	return suggestions
}

// GetWeeklySummary returns a summary of the weekly plan
func (s *WeeklyPlanningSession) GetWeeklySummary() string {
	var summary strings.Builder
	
	// Header
	summary.WriteString(fmt.Sprintf("Weekly Plan: %s - %s\n", 
		s.WeekStart.Format("Monday, Jan 2"), s.WeekEnd.Format("Friday, Jan 6, 2006")))
	summary.WriteString("\n")
	
	// Objectives
	if len(s.Objectives) > 0 {
		summary.WriteString("🎯 WEEKLY OBJECTIVES:\n")
		for i, obj := range s.Objectives {
			summary.WriteString(fmt.Sprintf("  %d. %s\n", i+1, obj.Title))
			if obj.Description != "" {
				summary.WriteString(fmt.Sprintf("     %s\n", obj.Description))
			}
		}
		summary.WriteString("\n")
	}
	
	// Strategic focus from journal
	if s.JournalEntry != "" {
		summary.WriteString("📝 STRATEGIC FOCUS:\n")
		// Truncate journal entry for summary
		journalSummary := s.JournalEntry
		if len(journalSummary) > 200 {
			journalSummary = journalSummary[:200] + "..."
		}
		summary.WriteString(fmt.Sprintf("  %s\n\n", journalSummary))
	}
	
	// Work streams
	if len(s.WorkStreams) > 0 {
		summary.WriteString("🚀 WORK STREAMS:\n")
		totalHours := 0.0
		for i, ws := range s.WorkStreams {
			summary.WriteString(fmt.Sprintf("  %d. %s (%.1fh)\n", i+1, ws.Title, ws.EstimatedHours))
			if ws.Description != "" {
				summary.WriteString(fmt.Sprintf("     %s\n", ws.Description))
			}
			if len(ws.Tasks) > 0 {
				summary.WriteString(fmt.Sprintf("     %d tasks included\n", len(ws.Tasks)))
			}
			totalHours += ws.EstimatedHours
		}
		summary.WriteString(fmt.Sprintf("\nTotal planned work: %.1f hours\n", totalHours))
		summary.WriteString("\n")
	}
	
	// Weekly capacity assessment
	weeklyCapacity := s.Context.WorkdayHours * 5 // 5 work days
	if len(s.WorkStreams) > 0 {
		totalPlanned := 0.0
		for _, ws := range s.WorkStreams {
			totalPlanned += ws.EstimatedHours
		}
		utilizationRate := totalPlanned / (weeklyCapacity * 0.7) // 70% for focused work
		
		if utilizationRate > 1.0 {
			summary.WriteString(fmt.Sprintf("⚠️ Weekly capacity: %.0f%% (may need to adjust scope)\n", utilizationRate*100))
		} else {
			summary.WriteString(fmt.Sprintf("✓ Weekly capacity: %.0f%% (good balance)\n", utilizationRate*100))
		}
	}
	
	return summary.String()
}

// GetJournalingPrompts returns prompts for strategic journaling
func (s *WeeklyPlanningSession) GetJournalingPrompts() []string {
	return []string{
		"What's most important to achieve this week?",
		"How do this week's objectives connect to your bigger goals?",
		"What obstacles might you face and how will you handle them?",
		"What support or resources do you need?",
		"What would make this week feel successful?",
		"How can you maintain balance and avoid burnout?",
		"What opportunities for growth or learning exist this week?",
	}
}