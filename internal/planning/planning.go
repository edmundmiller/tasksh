package planning

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// PlanningHorizon represents the time horizon for planning
type PlanningHorizon int

const (
	HorizonTomorrow PlanningHorizon = iota
	HorizonWeek
	HorizonQuick // Quick planning mode for busy mornings
)

// TaskCategory represents the priority category for staged planning
type TaskCategory int

const (
	CategoryCritical TaskCategory = iota // Must do today (urgency >20.0 or due today)
	CategoryImportant                    // Should do today (urgency 10.0-20.0)
	CategoryFlexible                     // Could do today (urgency <10.0)
)

// EnergyLevel represents the cognitive energy required for a task
type EnergyLevel int

const (
	EnergyHigh   EnergyLevel = iota // Requires focused, uninterrupted time
	EnergyMedium                    // Moderate cognitive load
	EnergyLow                       // Can be done with minimal focus
)

// PlannedTask represents a task with planning metadata
type PlannedTask struct {
	*taskwarrior.Task
	EstimatedHours   float64
	EstimationReason string
	Urgency          float64
	PlannedDate      time.Time
	IsScheduled      bool // true if task has a scheduled date
	IsDue            bool // true if task has a due date
	Category         TaskCategory // Critical/Important/Flexible categorization
	EnergyLevel      EnergyLevel  // Cognitive energy required
	OptimalTimeSlot  string       // Suggested time of day (e.g., "morning", "afternoon")
}

// PlanningSession represents a planning session
type PlanningSession struct {
	Horizon        PlanningHorizon
	Tasks          []PlannedTask // All tasks (for backward compatibility)
	
	// Categorized task lists for staged planning
	CriticalTasks  []PlannedTask
	ImportantTasks []PlannedTask
	FlexibleTasks  []PlannedTask
	BacklogTasks   []PlannedTask // Tasks not included in daily plan
	
	TotalHours     float64
	DailyCapacity  float64
	FocusCapacity  float64  // Realistic focused work capacity (typically 6h)
	WarningLevel   WarningLevel
	Date           time.Time // For daily planning, the target date
	
	// Smart limits
	MaxTasks       int     // Maximum tasks to include in daily plan
	MaxFocusHours  float64 // Maximum focused work hours
	BufferTime     float64 // Buffer percentage for interruptions
	
	timeDB         *timedb.TimeDB
}

// WarningLevel represents capacity warning levels
type WarningLevel int

const (
	WarningNone WarningLevel = iota
	WarningCaution  // 90%+ capacity
	WarningOverload // 100%+ capacity
)

// NewPlanningSession creates a new planning session
func NewPlanningSession(horizon PlanningHorizon) (*PlanningSession, error) {
	timeDB, err := timedb.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open time database: %w", err)
	}

	session := &PlanningSession{
		Horizon:       horizon,
		DailyCapacity: 8.0, // Default 8 hours per day
		FocusCapacity: 6.0, // Realistic focused work capacity
		MaxTasks:      8,   // Maximum tasks in daily plan
		MaxFocusHours: 6.0, // Maximum focused work hours
		BufferTime:    0.25, // 25% buffer for interruptions
		timeDB:        timeDB,
	}

	switch horizon {
	case HorizonTomorrow:
		session.Date = time.Now().AddDate(0, 0, 1) // Tomorrow
	case HorizonWeek:
		session.Date = time.Now() // Week starting today
	case HorizonQuick:
		session.Date = time.Now().AddDate(0, 0, 1) // Tomorrow
		session.MaxTasks = 3 // Limit for quick mode
		session.MaxFocusHours = 4.0 // Reduced for quick planning
	}

	return session, nil
}

// Close closes the planning session and releases resources
func (ps *PlanningSession) Close() error {
	if ps.timeDB != nil {
		return ps.timeDB.Close()
	}
	return nil
}

// LoadTasks loads and processes tasks for the planning horizon
func (ps *PlanningSession) LoadTasks() error {
	var uuids []string
	var err error

	switch ps.Horizon {
	case HorizonTomorrow:
		uuids, err = ps.getTasksForTomorrow()
	case HorizonWeek:
		uuids, err = ps.getTasksForWeek()
	case HorizonQuick:
		uuids, err = ps.getTasksForQuick()
	}

	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// Convert to PlannedTasks with metadata
	allTasks := make([]PlannedTask, 0, len(uuids))
	for _, uuid := range uuids {
		task, err := taskwarrior.GetTaskInfo(uuid)
		if err != nil {
			continue // Skip tasks we can't load
		}

		plannedTask := PlannedTask{
			Task:        task,
			PlannedDate: ps.Date,
		}

		// Calculate urgency
		plannedTask.Urgency = ps.calculateUrgency(task)

		// Determine if scheduled or due
		plannedTask.IsScheduled = task.Due != "" // Simplified - in real taskwarrior, there's a separate scheduled field
		plannedTask.IsDue = task.Due != ""

		// Get time estimation
		plannedTask.EstimatedHours, plannedTask.EstimationReason = ps.estimateTaskTime(task)

		// Categorize task based on urgency and due date
		plannedTask.Category = ps.categorizeTask(plannedTask)

		// Determine energy level and optimal time slot
		plannedTask.EnergyLevel = ps.calculateEnergyLevel(plannedTask)
		plannedTask.OptimalTimeSlot = ps.getOptimalTimeSlot(plannedTask)

		allTasks = append(allTasks, plannedTask)
	}

	// Sort all tasks by priority
	ps.sortTasksByPriority(allTasks)

	// Apply smart limits and organize into categories
	ps.organizeTasks(allTasks)

	// Calculate totals and warnings for the organized tasks
	ps.calculateTotals()

	return nil
}

// getTasksForTomorrow gets tasks relevant for tomorrow's planning
func (ps *PlanningSession) getTasksForTomorrow() ([]string, error) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrowStr := tomorrow.Format("2006-01-02")

	// Get tasks due tomorrow, scheduled tomorrow, or with high urgency
	filters := []string{
		fmt.Sprintf("(due:%s or urgency.over:15.0)", tomorrowStr),
		"and", "(+PENDING", "or", "+WAITING)",
	}

	return ps.executeTaskFilter(filters)
}

// getTasksForWeek gets tasks relevant for weekly planning
func (ps *PlanningSession) getTasksForWeek() ([]string, error) {
	endOfWeek := time.Now().AddDate(0, 0, 7)
	eowStr := endOfWeek.Format("2006-01-02")

	// Get tasks due this week or with moderate urgency
	filters := []string{
		fmt.Sprintf("(due.before:%s or urgency.over:10.0)", eowStr),
		"and", "(+PENDING", "or", "+WAITING)",
	}

	return ps.executeTaskFilter(filters)
}

// getTasksForQuick gets tasks for quick planning mode (only most critical)
func (ps *PlanningSession) getTasksForQuick() ([]string, error) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrowStr := tomorrow.Format("2006-01-02")

	// Get only the most critical tasks: due today/tomorrow or very high urgency
	filters := []string{
		fmt.Sprintf("(due:%s or due:today or urgency.over:25.0)", tomorrowStr),
		"and", "(+PENDING", "or", "+WAITING)",
	}

	return ps.executeTaskFilter(filters)
}

// categorizeTask determines the category (Critical/Important/Flexible) for a task
func (ps *PlanningSession) categorizeTask(task PlannedTask) TaskCategory {
	// Critical: Due today/tomorrow or very high urgency (>20.0)
	if task.IsDue || task.Urgency >= 20.0 {
		return CategoryCritical
	}
	
	// Important: Moderate to high urgency (10.0-20.0)
	if task.Urgency >= 10.0 {
		return CategoryImportant
	}
	
	// Everything else is flexible
	return CategoryFlexible
}

// sortTasksByPriority sorts tasks by urgency and due date
func (ps *PlanningSession) sortTasksByPriority(tasks []PlannedTask) {
	sort.Slice(tasks, func(i, j int) bool {
		taskA, taskB := tasks[i], tasks[j]

		// Due tasks first
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

// organizeTasks applies smart limits and organizes tasks into categories
func (ps *PlanningSession) organizeTasks(allTasks []PlannedTask) {
	// Initialize category slices
	ps.CriticalTasks = []PlannedTask{}
	ps.ImportantTasks = []PlannedTask{}
	ps.FlexibleTasks = []PlannedTask{}
	ps.BacklogTasks = []PlannedTask{}

	var totalHours float64
	var taskCount int

	// First pass: Add critical tasks (always include, but limit hours)
	for _, task := range allTasks {
		if task.Category == CategoryCritical {
			if totalHours + task.EstimatedHours <= ps.MaxFocusHours || len(ps.CriticalTasks) < 3 {
				ps.CriticalTasks = append(ps.CriticalTasks, task)
				totalHours += task.EstimatedHours
				taskCount++
			} else {
				ps.BacklogTasks = append(ps.BacklogTasks, task)
			}
		}
	}

	// Second pass: Add important tasks until we hit limits
	for _, task := range allTasks {
		if task.Category == CategoryImportant {
			if taskCount < ps.MaxTasks && totalHours + task.EstimatedHours <= ps.MaxFocusHours {
				ps.ImportantTasks = append(ps.ImportantTasks, task)
				totalHours += task.EstimatedHours
				taskCount++
			} else {
				ps.BacklogTasks = append(ps.BacklogTasks, task)
			}
		}
	}

	// Third pass: Add flexible tasks if there's still capacity
	for _, task := range allTasks {
		if task.Category == CategoryFlexible {
			if taskCount < ps.MaxTasks && totalHours + task.EstimatedHours <= ps.MaxFocusHours {
				ps.FlexibleTasks = append(ps.FlexibleTasks, task)
				totalHours += task.EstimatedHours
				taskCount++
			} else {
				ps.BacklogTasks = append(ps.BacklogTasks, task)
			}
		}
	}

	// Create the combined Tasks slice for backward compatibility
	ps.Tasks = make([]PlannedTask, 0, len(ps.CriticalTasks)+len(ps.ImportantTasks)+len(ps.FlexibleTasks))
	ps.Tasks = append(ps.Tasks, ps.CriticalTasks...)
	ps.Tasks = append(ps.Tasks, ps.ImportantTasks...)
	ps.Tasks = append(ps.Tasks, ps.FlexibleTasks...)
}

// calculateEnergyLevel determines the cognitive energy required for a task
func (ps *PlanningSession) calculateEnergyLevel(task PlannedTask) EnergyLevel {
	// High energy indicators
	highEnergyKeywords := []string{"design", "code", "develop", "write", "create", "analyze", "research", "plan", "architect", "review"}
	
	// Low energy indicators  
	lowEnergyKeywords := []string{"email", "call", "meeting", "standup", "sync", "update", "check", "status", "admin", "file"}
	
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
	
	// Consider task duration and complexity
	if task.EstimatedHours >= 2.0 {
		return EnergyHigh // Long tasks usually need focus
	} else if task.EstimatedHours <= 0.5 {
		return EnergyLow // Quick tasks are usually low energy
	}
	
	// Consider priority as energy indicator
	if task.Priority == "H" {
		return EnergyHigh
	}
	
	// Default to medium energy
	return EnergyMedium
}

// getOptimalTimeSlot suggests the best time of day for a task
func (ps *PlanningSession) getOptimalTimeSlot(task PlannedTask) string {
	switch task.EnergyLevel {
	case EnergyHigh:
		return "morning" // High focus work best in morning
	case EnergyMedium:
		// Spread medium tasks throughout the day
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

// executeTaskFilter executes a taskwarrior filter and returns UUIDs
func (ps *PlanningSession) executeTaskFilter(filters []string) ([]string, error) {
	// Use the existing taskwarrior package approach
	args := append([]string{
		"rc.color=off",
		"rc.detection=off",
		"rc._forcecolor=off",
		"rc.verbose=nothing",
	}, filters...)
	args = append(args, "uuids")

	output, err := executeTask(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute task filter: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Fields(output), nil
}

// executeTask is a helper to run taskwarrior commands (copied from taskwarrior package pattern)
func executeTask(args ...string) (string, error) {
	cmd := exec.Command("task", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("task command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// calculateUrgency calculates urgency score for a task
func (ps *PlanningSession) calculateUrgency(task *taskwarrior.Task) float64 {
	// Simplified urgency calculation
	// In real implementation, this would use taskwarrior's urgency calculation
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
		// Parse due date and calculate urgency based on proximity
		if dueTime, err := time.Parse("2006-01-02T15:04:05Z", task.Due); err == nil {
			daysUntilDue := time.Until(dueTime).Hours() / 24
			if daysUntilDue <= 1 {
				urgency += 12.0
			} else if daysUntilDue <= 7 {
				urgency += 6.0
			} else if daysUntilDue <= 30 {
				urgency += 2.0
			}
		}
	}

	// Project contribution (projects are important)
	if task.Project != "" {
		urgency += 1.0
	}

	return urgency
}

// estimateTaskTime estimates time for a task using historical data
func (ps *PlanningSession) estimateTaskTime(task *taskwarrior.Task) (float64, string) {
	if ps.timeDB == nil {
		return 2.0, "Default estimate (no historical data)"
	}

	estimate, reason, err := ps.timeDB.EstimateTimeForTask(task)
	if err != nil || estimate <= 0 {
		// Fallback estimates based on priority/complexity
		switch task.Priority {
		case "H":
			return 3.0, "High priority task estimate"
		case "M":
			return 2.0, "Medium priority task estimate"
		case "L":
			return 1.0, "Low priority task estimate"
		default:
			return 2.0, "Default task estimate"
		}
	}

	// Add 15% buffer to estimates
	bufferedEstimate := estimate * 1.15
	return bufferedEstimate, reason + " (with 15% buffer)"
}

// sortTasks sorts tasks by planning priority
func (ps *PlanningSession) sortTasks() {
	sort.Slice(ps.Tasks, func(i, j int) bool {
		taskA, taskB := ps.Tasks[i], ps.Tasks[j]

		// Due tasks first
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

// calculateTotals calculates total hours and warning levels
func (ps *PlanningSession) calculateTotals() {
	ps.TotalHours = 0.0
	for _, task := range ps.Tasks {
		ps.TotalHours += task.EstimatedHours
	}

	// Determine warning level based on realistic focus capacity
	focusRatio := ps.TotalHours / ps.FocusCapacity
	if focusRatio >= 1.0 {
		ps.WarningLevel = WarningOverload
	} else if focusRatio >= 0.9 {
		ps.WarningLevel = WarningCaution
	} else {
		ps.WarningLevel = WarningNone
	}
}

// GetCapacityStatus returns a human-readable capacity status with specific guidance
func (ps *PlanningSession) GetCapacityStatus() string {
	used := ps.TotalHours
	available := ps.FocusCapacity
	percentage := int((used / available) * 100)

	switch ps.WarningLevel {
	case WarningOverload:
		overload := used - available
		suggestion := ps.getOverloadSuggestion()
		return fmt.Sprintf("⚠️ Overloaded by %.1fh (%d%% of focus capacity) - %s", overload, percentage, suggestion)
	case WarningCaution:
		return fmt.Sprintf("⚠️ Near capacity: %.1fh/%.1fh (%d%%) - Consider deferring flexible tasks", used, available, percentage)
	default:
		taskCount := len(ps.CriticalTasks) + len(ps.ImportantTasks) + len(ps.FlexibleTasks)
		return fmt.Sprintf("%.1fh/%.1fh (%d%% focus capacity, %d tasks)", used, available, percentage, taskCount)
	}
}

// getOverloadSuggestion provides specific guidance when overloaded
func (ps *PlanningSession) getOverloadSuggestion() string {
	if len(ps.FlexibleTasks) > 0 {
		return "Move flexible tasks to backlog"
	} else if len(ps.ImportantTasks) > 2 {
		return "Consider deferring some important tasks"
	} else {
		return "Break down large tasks or defer to tomorrow"
	}
}

// MoveTask moves a task to a different position in the playlist
func (ps *PlanningSession) MoveTask(fromIndex, toIndex int) error {
	if fromIndex < 0 || fromIndex >= len(ps.Tasks) || toIndex < 0 || toIndex >= len(ps.Tasks) {
		return fmt.Errorf("invalid task indices")
	}

	if fromIndex == toIndex {
		return nil
	}

	// Create a new slice with the task moved
	task := ps.Tasks[fromIndex]
	newTasks := make([]PlannedTask, 0, len(ps.Tasks))
	
	// Add tasks up to the target position
	for i := 0; i < len(ps.Tasks); i++ {
		if i == fromIndex {
			continue // Skip the task we're moving
		}
		if len(newTasks) == toIndex {
			newTasks = append(newTasks, task) // Insert the moved task
		}
		newTasks = append(newTasks, ps.Tasks[i])
	}
	
	// If target position is at the end, append the moved task
	if toIndex >= len(newTasks) {
		newTasks = append(newTasks, task)
	}
	
	ps.Tasks = newTasks
	return nil
}

// RemoveTask removes a task from the planning session
func (ps *PlanningSession) RemoveTask(index int) error {
	if index < 0 || index >= len(ps.Tasks) {
		return fmt.Errorf("invalid task index")
	}

	ps.Tasks = append(ps.Tasks[:index], ps.Tasks[index+1:]...)
	ps.calculateTotals()
	return nil
}

// GetProjectedCompletionTimes calculates when tasks would be completed based on order
func (ps *PlanningSession) GetProjectedCompletionTimes(startTime time.Time) []time.Time {
	completionTimes := make([]time.Time, len(ps.Tasks))
	currentTime := startTime

	for i, task := range ps.Tasks {
		currentTime = currentTime.Add(time.Duration(task.EstimatedHours * float64(time.Hour)))
		completionTimes[i] = currentTime
	}

	return completionTimes
}