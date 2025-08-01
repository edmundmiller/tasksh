package shared

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// TaskAnalyzer provides utilities for analyzing and categorizing tasks
type TaskAnalyzer struct {
	timeDB *timedb.TimeDB
}

// NewTaskAnalyzer creates a new task analyzer
func NewTaskAnalyzer(timeDB *timedb.TimeDB) *TaskAnalyzer {
	return &TaskAnalyzer{timeDB: timeDB}
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
	if ta.timeDB == nil {
		return ta.fallbackEstimate(task), "Default estimate (no historical data)"
	}

	estimate, reason, err := ta.timeDB.EstimateTimeForTask(task)
	if err != nil || estimate <= 0 {
		return ta.fallbackEstimate(task), "Fallback estimate"
	}

	// Add 15% buffer to estimates
	bufferedEstimate := estimate * 1.15
	return bufferedEstimate, reason + " (with 15% buffer)"
}

// fallbackEstimate provides fallback time estimates
func (ta *TaskAnalyzer) fallbackEstimate(task *taskwarrior.Task) float64 {
	switch task.Priority {
	case "H":
		return 3.0
	case "M":
		return 2.0
	case "L":
		return 1.0
	default:
		return 2.0
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