package daily

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/planning/shared"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// WorkflowStep represents a step in the daily planning workflow
type WorkflowStep int

// BackgroundLoadResult represents the result of background task loading
type BackgroundLoadResult struct {
	Error error
}

const (
	StepReflection WorkflowStep = iota
	StepTaskSelection
	StepWorkloadAssessment
	StepFinalization
	StepSummary
	StepCompleted
)

// String returns a human-readable step name
func (s WorkflowStep) String() string {
	switch s {
	case StepReflection:
		return "Reflect on Yesterday"
	case StepTaskSelection:
		return "Select Today's Tasks"
	case StepWorkloadAssessment:
		return "Assess Workload"
	case StepFinalization:
		return "Finalize Plan"
	case StepSummary:
		return "Review Plan"
	case StepCompleted:
		return "Planning Complete"
	default:
		return "Unknown Step"
	}
}

// DailyPlanningSession represents a guided daily planning session
type DailyPlanningSession struct {
	Context        shared.PlanningContext
	CurrentStep    WorkflowStep
	Reflection     *shared.ReflectionData
	AvailableTasks []shared.PlannedTask
	SelectedTasks  []shared.PlannedTask
	Assessment     *shared.WorkloadAssessment
	DailyFocus     string // Main theme/intention for the day
	
	// Session metadata
	StartedAt   time.Time
	CompletedAt *time.Time
	
	// Internal utilities
	analyzer *shared.TaskAnalyzer
	
	// Background loading state
	tasksLoaded bool
	loadError   error
	
	// Existing scheduled tasks
	ExistingScheduledTasks []shared.PlannedTask
	HasExistingPlan       bool
}

// NewDailyPlanningSession creates a new daily planning session
func NewDailyPlanningSession(targetDate time.Time) (*DailyPlanningSession, error) {
	timeDB, err := timedb.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open time database: %w", err)
	}

	context := shared.PlanningContext{
		Date:         targetDate,
		TimeDB:       timeDB,
		WorkdayHours: 8.0, // Default 8 hours
		EnergyLevel:  shared.EnergyMedium, // Default, will be updated during workflow
	}

	session := &DailyPlanningSession{
		Context:     context,
		CurrentStep: StepReflection,
		StartedAt:   time.Now(),
		analyzer:    shared.NewTaskAnalyzer(timeDB),
	}
	
	// Check for existing scheduled tasks
	scheduledTasks, err := session.CheckScheduledTasks()
	if err != nil {
		// Log but don't fail - we can continue without this check
		fmt.Printf("Warning: Could not check scheduled tasks: %v\n", err)
	} else if len(scheduledTasks) > 0 {
		session.ExistingScheduledTasks = scheduledTasks
		session.HasExistingPlan = true
		// Pre-select the scheduled tasks
		session.SelectedTasks = scheduledTasks
	}

	return session, nil
}

// Close closes the planning session and releases resources
func (s *DailyPlanningSession) Close() error {
	if s.Context.TimeDB != nil {
		return s.Context.TimeDB.Close()
	}
	return nil
}

// GetStepProgress returns the current step and total steps
func (s *DailyPlanningSession) GetStepProgress() (current int, total int) {
	// StepCompleted is not an actual step, so total steps is StepSummary + 1
	return int(s.CurrentStep), int(StepSummary) + 1
}

// IsCompleted returns true if the planning session is completed
func (s *DailyPlanningSession) IsCompleted() bool {
	return s.CurrentStep == StepCompleted
}

// NextStep advances to the next step in the workflow
func (s *DailyPlanningSession) NextStep() error {
	switch s.CurrentStep {
	case StepReflection:
		// Only load if not already loaded in background
		if !s.tasksLoaded {
			if err := s.loadAvailableTasks(); err != nil {
				return fmt.Errorf("failed to load tasks: %w", err)
			}
		} else if s.loadError != nil {
			// If background loading failed, return that error
			return fmt.Errorf("failed to load tasks: %w", s.loadError)
		}
		s.CurrentStep = StepTaskSelection
	case StepTaskSelection:
		s.CurrentStep = StepWorkloadAssessment
	case StepWorkloadAssessment:
		s.CurrentStep = StepFinalization
	case StepFinalization:
		s.CurrentStep = StepSummary
	case StepSummary:
		s.CurrentStep = StepCompleted
		completedAt := time.Now()
		s.CompletedAt = &completedAt
	default:
		return fmt.Errorf("cannot advance from current step: %s", s.CurrentStep)
	}
	return nil
}

// PreviousStep goes back to the previous step (if allowed)
func (s *DailyPlanningSession) PreviousStep() error {
	switch s.CurrentStep {
	case StepTaskSelection:
		s.CurrentStep = StepReflection
	case StepWorkloadAssessment:
		s.CurrentStep = StepTaskSelection
	case StepFinalization:
		s.CurrentStep = StepWorkloadAssessment
	case StepSummary:
		s.CurrentStep = StepFinalization
	default:
		return fmt.Errorf("cannot go back from current step: %s", s.CurrentStep)
	}
	return nil
}

// SetReflectionData sets the reflection data for yesterday
func (s *DailyPlanningSession) SetReflectionData(reflection *shared.ReflectionData) {
	s.Reflection = reflection
	// Update energy level based on reflection
	if reflection != nil {
		s.Context.EnergyLevel = reflection.EnergyLevel
	}
}

// GetReflectionPrompts returns prompts for yesterday's reflection
func (s *DailyPlanningSession) GetReflectionPrompts() []string {
	yesterday := s.Context.Date.AddDate(0, 0, -1)
	return []string{
		fmt.Sprintf("What did you accomplish yesterday (%s)?", yesterday.Format("Monday, Jan 2")),
		"What went well with your work?",
		"What challenges did you face?",
		"How was your energy level throughout the day?",
		"What would you do differently?",
	}
}

// CheckScheduledTasks checks for tasks already scheduled for the target date
func (s *DailyPlanningSession) CheckScheduledTasks() ([]shared.PlannedTask, error) {
	dateStr := s.Context.Date.Format("2006-01-02")
	
	// Get tasks scheduled for today
	filters := []string{
		fmt.Sprintf("scheduled:%s", dateStr),
		"and", "(+PENDING", "or", "+WAITING)",
	}

	uuids, err := shared.LoadTasksByFilter(filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to load scheduled tasks: %w", err)
	}

	if len(uuids) == 0 {
		return []shared.PlannedTask{}, nil
	}

	// Batch load all tasks
	taskMap, err := taskwarrior.BatchLoadTasks(uuids)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load scheduled tasks: %w", err)
	}

	// Convert to PlannedTasks
	scheduledTasks := make([]shared.PlannedTask, 0, len(uuids))
	for _, uuid := range uuids {
		task, ok := taskMap[uuid]
		if !ok {
			continue
		}

		plannedTask := s.analyzer.AnalyzeTask(task, s.Context.Date)
		scheduledTasks = append(scheduledTasks, plannedTask)
	}

	return scheduledTasks, nil
}

// loadAvailableTasks loads tasks relevant for today's planning
func (s *DailyPlanningSession) loadAvailableTasks() error {
	todayStr := s.Context.Date.Format("2006-01-02")
	
	// Get tasks due today, overdue, or with high urgency
	filters := []string{
		fmt.Sprintf("(due:%s or due.before:%s or urgency.over:15.0)", todayStr, todayStr),
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

		plannedTask := s.analyzer.AnalyzeTask(task, s.Context.Date)
		s.AvailableTasks = append(s.AvailableTasks, plannedTask)
	}

	// Sort by priority
	shared.SortTasksByPriority(s.AvailableTasks)

	// Mark as loaded
	s.tasksLoaded = true
	return nil
}

// AddTaskToSelection adds a task to today's plan
func (s *DailyPlanningSession) AddTaskToSelection(taskIndex int) error {
	if taskIndex < 0 || taskIndex >= len(s.AvailableTasks) {
		return fmt.Errorf("invalid task index: %d", taskIndex)
	}

	task := s.AvailableTasks[taskIndex]
	
	// Check if already selected
	for _, selected := range s.SelectedTasks {
		if selected.UUID == task.UUID {
			return fmt.Errorf("task already selected")
		}
	}

	s.SelectedTasks = append(s.SelectedTasks, task)
	return nil
}

// RemoveTaskFromSelection removes a task from today's plan
func (s *DailyPlanningSession) RemoveTaskFromSelection(taskIndex int) error {
	if taskIndex < 0 || taskIndex >= len(s.SelectedTasks) {
		return fmt.Errorf("invalid task index: %d", taskIndex)
	}

	s.SelectedTasks = append(s.SelectedTasks[:taskIndex], s.SelectedTasks[taskIndex+1:]...)
	return nil
}

// CalculateWorkloadAssessment calculates the workload assessment for selected tasks
func (s *DailyPlanningSession) CalculateWorkloadAssessment() *shared.WorkloadAssessment {
	_, _, energyBreakdown := shared.CalculateWorkloadSummary(s.SelectedTasks)
	
	// Calculate realistic focus hours (typically 60-70% of workday)
	focusHours := s.Context.WorkdayHours * 0.65
	
	// Calculate meeting/admin time (rough estimate)
	meetingHours := s.Context.WorkdayHours * 0.2
	
	assessment := &shared.WorkloadAssessment{
		AvailableHours:     s.Context.WorkdayHours,
		FocusHours:         focusHours,
		MeetingHours:       meetingHours,
		BufferPercentage:   0.15, // 15% buffer for interruptions
		EnergyDistribution: energyBreakdown,
	}

	s.Assessment = assessment
	return assessment
}

// GetCapacityWarning returns a warning if the workload is too high
func (s *DailyPlanningSession) GetCapacityWarning() string {
	if s.Assessment == nil {
		return ""
	}

	totalHours, _, _ := shared.CalculateWorkloadSummary(s.SelectedTasks)
	utilizationRate := totalHours / s.Assessment.FocusHours

	if utilizationRate >= 1.2 {
		overload := totalHours - s.Assessment.FocusHours
		return fmt.Sprintf("⚠️ Significantly overloaded by %.1fh (%.0f%% capacity) - Consider deferring flexible tasks", 
			overload, utilizationRate*100)
	} else if utilizationRate >= 1.0 {
		overload := totalHours - s.Assessment.FocusHours
		return fmt.Sprintf("⚠️ Overloaded by %.1fh (%.0f%% capacity) - Consider reducing scope", 
			overload, utilizationRate*100)
	} else if utilizationRate >= 0.9 {
		return fmt.Sprintf("⚠️ Near capacity (%.0f%%) - Plan looks full", utilizationRate*100)
	}

	return fmt.Sprintf("✓ Good capacity (%.0f%% of focus time)", utilizationRate*100)
}

// SetDailyFocus sets the main focus/intention for the day
func (s *DailyPlanningSession) SetDailyFocus(focus string) {
	s.DailyFocus = strings.TrimSpace(focus)
}

// StartBackgroundTaskLoading starts loading tasks in the background
// Returns a tea.Cmd that performs the loading
func (s *DailyPlanningSession) StartBackgroundTaskLoading() func() tea.Msg {
	return func() tea.Msg {
		// Load tasks in background
		err := s.loadAvailableTasks()
		if err != nil {
			s.loadError = err
		}
		return BackgroundLoadResult{Error: err}
	}
}

// IsTasksLoaded returns whether tasks have been loaded
func (s *DailyPlanningSession) IsTasksLoaded() bool {
	return s.tasksLoaded
}

// GetLoadError returns any error from background loading
func (s *DailyPlanningSession) GetLoadError() error {
	return s.loadError
}

// ScheduleSelectedTasks marks all selected tasks as scheduled for the target date
func (s *DailyPlanningSession) ScheduleSelectedTasks() error {
	dateStr := s.Context.Date.Format("2006-01-02")
	
	for _, task := range s.SelectedTasks {
		err := taskwarrior.ModifyTask(task.UUID, fmt.Sprintf("scheduled:%s", dateStr))
		if err != nil {
			return fmt.Errorf("failed to schedule task %s: %w", task.UUID, err)
		}
	}
	
	// Mark daily planning as complete
	if err := s.markDailyPlanningComplete(); err != nil {
		// Log but don't fail - this is a nice-to-have feature
		fmt.Printf("Warning: Could not mark daily planning as complete: %v\n", err)
	}
	
	return nil
}

// markDailyPlanningComplete creates or updates a recurring task to track daily planning completion
func (s *DailyPlanningSession) markDailyPlanningComplete() error {
	// First, check if there's already a daily planning task for today
	filters := []string{
		fmt.Sprintf("description:'Daily Planning - %s'", s.Context.Date.Format("January 2, 2006")),
		"(+PENDING or +COMPLETED)",
	}
	
	existingUUIDs, err := shared.LoadTasksByFilter(filters...)
	if err != nil {
		return fmt.Errorf("failed to check for existing daily planning task: %w", err)
	}
	
	if len(existingUUIDs) > 0 {
		// Mark the existing task as completed
		for _, uuid := range existingUUIDs {
			if err := taskwarrior.CompleteTask(uuid); err != nil {
				return fmt.Errorf("failed to complete daily planning task: %w", err)
			}
		}
	} else {
		// Create a new task for today's planning and immediately complete it
		description := fmt.Sprintf("Daily Planning - %s", s.Context.Date.Format("January 2, 2006"))
		project := "+planning"
		tags := []string{"+daily", "+ritual"}
		
		// Create the task
		uuid, err := taskwarrior.AddTask(description, project, tags...)
		if err != nil {
			return fmt.Errorf("failed to create daily planning task: %w", err)
		}
		
		// Complete it immediately
		if err := taskwarrior.CompleteTask(uuid); err != nil {
			return fmt.Errorf("failed to complete daily planning task: %w", err)
		}
	}
	
	// Also ensure there's a recurring task for tomorrow's planning
	if err := s.ensureRecurringPlanningTask(); err != nil {
		return fmt.Errorf("failed to ensure recurring planning task: %w", err)
	}
	
	return nil
}

// ensureRecurringPlanningTask ensures there's a recurring task for daily planning
func (s *DailyPlanningSession) ensureRecurringPlanningTask() error {
	// Check if recurring task exists
	filters := []string{
		"description:'Daily Planning'",
		"+PENDING",
		"recur.any:",
	}
	
	recurringUUIDs, err := shared.LoadTasksByFilter(filters...)
	if err != nil {
		return fmt.Errorf("failed to check for recurring planning task: %w", err)
	}
	
	// If no recurring task exists, create one
	if len(recurringUUIDs) == 0 {
		description := "Daily Planning"
		project := "+planning"
		tags := []string{"+daily", "+ritual", "recur:daily", "due:tomorrow", "scheduled:tomorrow"}
		
		_, err := taskwarrior.AddTask(description, project, tags...)
		if err != nil {
			return fmt.Errorf("failed to create recurring daily planning task: %w", err)
		}
	}
	
	return nil
}

// GetDailySummary returns a summary of the planned day
func (s *DailyPlanningSession) GetDailySummary() string {
	if len(s.SelectedTasks) == 0 {
		return "No tasks planned for today."
	}

	var summary strings.Builder
	
	// Header
	summary.WriteString(fmt.Sprintf("Daily Plan for %s\n", s.Context.Date.Format("Monday, January 2, 2006")))
	
	if s.DailyFocus != "" {
		summary.WriteString(fmt.Sprintf("Focus: %s\n", s.DailyFocus))
	}
	
	summary.WriteString("\n")

	// Tasks by category
	criticalTasks := []shared.PlannedTask{}
	importantTasks := []shared.PlannedTask{}
	flexibleTasks := []shared.PlannedTask{}

	for _, task := range s.SelectedTasks {
		switch task.Category {
		case shared.CategoryCritical:
			criticalTasks = append(criticalTasks, task)
		case shared.CategoryImportant:
			importantTasks = append(importantTasks, task)
		case shared.CategoryFlexible:
			flexibleTasks = append(flexibleTasks, task)
		}
	}

	// Render sections
	if len(criticalTasks) > 0 {
		summary.WriteString("CRITICAL TASKS:\n")
		for i, task := range criticalTasks {
			summary.WriteString(fmt.Sprintf("  %d. %s (%.1fh)\n", i+1, task.Description, task.EstimatedHours))
		}
		summary.WriteString("\n")
	}

	if len(importantTasks) > 0 {
		summary.WriteString("IMPORTANT TASKS:\n")
		for i, task := range importantTasks {
			summary.WriteString(fmt.Sprintf("  %d. %s (%.1fh)\n", i+1, task.Description, task.EstimatedHours))
		}
		summary.WriteString("\n")
	}

	if len(flexibleTasks) > 0 {
		summary.WriteString("FLEXIBLE TASKS:\n")
		for i, task := range flexibleTasks {
			summary.WriteString(fmt.Sprintf("  %d. %s (%.1fh)\n", i+1, task.Description, task.EstimatedHours))
		}
		summary.WriteString("\n")
	}

	// Summary stats
	totalHours, _, _ := shared.CalculateWorkloadSummary(s.SelectedTasks)
	summary.WriteString(fmt.Sprintf("Total: %d tasks, %.1f hours planned\n", len(s.SelectedTasks), totalHours))
	
	if s.Assessment != nil {
		summary.WriteString(s.GetCapacityWarning())
	}

	return summary.String()
}