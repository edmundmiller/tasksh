package shared

import (
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// PlanningContext represents the context for any planning session
type PlanningContext struct {
	UserID      string    // For future multi-user support
	Date        time.Time // The target date for planning
	TimeDB      *timedb.TimeDB
	WorkdayHours float64  // Available working hours per day
	EnergyLevel  EnergyLevel
}

// EnergyLevel represents the user's current energy and focus capacity
type EnergyLevel int

const (
	EnergyLow EnergyLevel = iota
	EnergyMedium
	EnergyHigh
)

// String returns a human-readable energy level
func (e EnergyLevel) String() string {
	switch e {
	case EnergyLow:
		return "Low energy - minimal focus work"
	case EnergyMedium:
		return "Medium energy - balanced work"
	case EnergyHigh:
		return "High energy - deep focus work"
	default:
		return "Unknown energy level"
	}
}

// TaskCategory represents different priority levels for planning
type TaskCategory int

const (
	CategoryCritical TaskCategory = iota // Must do (urgent/due)
	CategoryImportant                    // Should do (important)
	CategoryFlexible                     // Could do (optional)
)

// String returns a human-readable category name
func (c TaskCategory) String() string {
	switch c {
	case CategoryCritical:
		return "Critical"
	case CategoryImportant:
		return "Important"
	case CategoryFlexible:
		return "Flexible"
	default:
		return "Unknown"
	}
}

// PlannedTask represents a task with planning metadata
type PlannedTask struct {
	*taskwarrior.Task
	EstimatedHours   float64
	EstimationReason string
	Urgency          float64
	Category         TaskCategory
	OptimalTimeSlot  string // morning, afternoon, evening, anytime
	RequiredEnergy   EnergyLevel
	PlannedDate      time.Time
	IsScheduled      bool
	IsDue           bool
}

// ReflectionData represents reflection information from previous work
type ReflectionData struct {
	Date             time.Time
	CompletedTasks   []CompletedTaskSummary
	TotalHoursWorked float64
	EnergyLevel      EnergyLevel
	Accomplishments  []string
	Challenges       []string
	Lessons          []string
	Mood             string
}

// CompletedTaskSummary represents a summary of completed work
type CompletedTaskSummary struct {
	UUID         string
	Description  string
	TimeSpent    float64
	CompletedAt  time.Time
	SatisfactionLevel int // 1-5 scale
}

// Objective represents a high-level strategic goal
type Objective struct {
	ID           string
	Title        string
	Description  string
	SuccessCriteria []string
	EstimatedWeeks  int
	Priority     int
	CreatedAt    time.Time
	TargetDate   *time.Time
	Status       ObjectiveStatus
}

// ObjectiveStatus represents the status of an objective
type ObjectiveStatus int

const (
	ObjectiveActive ObjectiveStatus = iota
	ObjectiveCompleted
	ObjectiveDeferred
	ObjectiveCancelled
)

// WorkloadAssessment represents the user's capacity analysis
type WorkloadAssessment struct {
	AvailableHours    float64
	FocusHours        float64  // Realistic deep work capacity
	MeetingHours      float64  // Time blocked for meetings
	BufferPercentage  float64  // Buffer for interruptions (0.0-1.0)
	EnergyDistribution map[EnergyLevel]float64 // How energy is distributed across day
	OptimalTimeBlocks []TimeBlock
}

// TimeBlock represents an optimal time slot for specific types of work
type TimeBlock struct {
	StartTime   time.Time
	EndTime     time.Time
	EnergyLevel EnergyLevel
	Description string
}

// PlanningSession represents common planning session data
type PlanningSession struct {
	ID           string
	Type         string // "daily" or "weekly"
	Context      PlanningContext
	CreatedAt    time.Time
	CompletedAt  *time.Time
	Tasks        []PlannedTask
	Reflection   *ReflectionData
	Assessment   *WorkloadAssessment
}