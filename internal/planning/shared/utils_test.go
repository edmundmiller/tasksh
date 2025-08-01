package shared

import (
	"testing"
	"time"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestEnergyLevel_String(t *testing.T) {
	tests := []struct {
		level    EnergyLevel
		expected string
	}{
		{EnergyLow, "Low energy - minimal focus work"},
		{EnergyMedium, "Medium energy - balanced work"},
		{EnergyHigh, "High energy - deep focus work"},
		{EnergyLevel(999), "Unknown energy level"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("EnergyLevel(%d).String() = %s, want %s", int(tt.level), result, tt.expected)
		}
	}
}

func TestTaskCategory_String(t *testing.T) {
	tests := []struct {
		category TaskCategory
		expected string
	}{
		{CategoryCritical, "Critical"},
		{CategoryImportant, "Important"},
		{CategoryFlexible, "Flexible"},
		{TaskCategory(999), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.category.String()
		if result != tt.expected {
			t.Errorf("TaskCategory(%d).String() = %s, want %s", int(tt.category), result, tt.expected)
		}
	}
}

func TestNewTaskAnalyzer(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	if analyzer == nil {
		t.Fatal("Expected task analyzer to be created")
	}

	if analyzer.timeDB != nil {
		t.Error("Expected timeDB to be nil when passed nil")
	}
}

func TestCalculateUrgency(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	tests := []struct {
		name     string
		task     *taskwarrior.Task
		expected float64
	}{
		{
			name:     "High priority task",
			task:     &taskwarrior.Task{Priority: "H"},
			expected: 6.0,
		},
		{
			name:     "Medium priority task",
			task:     &taskwarrior.Task{Priority: "M"},
			expected: 1.8,
		},
		{
			name:     "Low priority task",
			task:     &taskwarrior.Task{Priority: "L"},
			expected: 0.0,
		},
		{
			name:     "Task with project",
			task:     &taskwarrior.Task{Project: "TestProject"},
			expected: 1.0,
		},
		{
			name:     "Task due today",
			task:     &taskwarrior.Task{Due: time.Now().Add(12*time.Hour).Format("2006-01-02T15:04:05Z")},
			expected: 12.0, // Due today/tomorrow gets 12.0 urgency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateUrgency(tt.task)
			if result != tt.expected {
				t.Errorf("calculateUrgency() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestCategorizeTask(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	tests := []struct {
		name     string
		task     PlannedTask
		expected TaskCategory
	}{
		{
			name:     "Due task should be critical",
			task:     PlannedTask{IsDue: true, Urgency: 5.0},
			expected: CategoryCritical,
		},
		{
			name:     "High urgency task should be critical",
			task:     PlannedTask{IsDue: false, Urgency: 25.0},
			expected: CategoryCritical,
		},
		{
			name:     "Medium urgency task should be important",
			task:     PlannedTask{IsDue: false, Urgency: 15.0},
			expected: CategoryImportant,
		},
		{
			name:     "Low urgency task should be flexible",
			task:     PlannedTask{IsDue: false, Urgency: 5.0},
			expected: CategoryFlexible,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.categorizeTask(tt.task)
			if result != tt.expected {
				t.Errorf("categorizeTask() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsDueToday(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	targetDate := time.Date(2023, 12, 15, 10, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		dueDate  string
		expected bool
	}{
		{
			name:     "Task due today",
			dueDate:  "2023-12-15T15:30:00Z",
			expected: true,
		},
		{
			name:     "Task due tomorrow",
			dueDate:  "2023-12-16T09:00:00Z",
			expected: false,
		},
		{
			name:     "Task due yesterday",
			dueDate:  "2023-12-14T18:00:00Z",
			expected: false,
		},
		{
			name:     "Task with no due date",
			dueDate:  "",
			expected: false,
		},
		{
			name:     "Task with invalid due date",
			dueDate:  "invalid-date",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &taskwarrior.Task{Due: tt.dueDate}
			result := analyzer.isDueToday(task, targetDate)
			if result != tt.expected {
				t.Errorf("isDueToday() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateRequiredEnergy(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	tests := []struct {
		name        string
		description string
		priority    string
		expected    EnergyLevel
	}{
		{
			name:        "High energy keyword",
			description: "Design new architecture",
			expected:    EnergyHigh,
		},
		{
			name:        "Low energy keyword",
			description: "Send email to team",
			expected:    EnergyLow,
		},
		{
			name:        "High priority task",
			description: "Regular task",
			priority:    "H",
			expected:    EnergyHigh,
		},
		{
			name:        "Regular task",
			description: "Some regular work",
			expected:    EnergyMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &taskwarrior.Task{
				Description: tt.description,
				Priority:    tt.priority,
			}
			result := analyzer.calculateRequiredEnergy(task)
			if result != tt.expected {
				t.Errorf("calculateRequiredEnergy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetOptimalTimeSlot(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	tests := []struct {
		name          string
		requiredEnergy EnergyLevel
		category      TaskCategory
		expected      string
	}{
		{
			name:          "High energy task",
			requiredEnergy: EnergyHigh,
			expected:      "morning",
		},
		{
			name:          "Medium energy critical task",
			requiredEnergy: EnergyMedium,
			category:      CategoryCritical,
			expected:      "morning",
		},
		{
			name:          "Medium energy important task",
			requiredEnergy: EnergyMedium,
			category:      CategoryImportant,
			expected:      "afternoon",
		},
		{
			name:          "Low energy task",
			requiredEnergy: EnergyLow,
			expected:      "anytime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := PlannedTask{
				RequiredEnergy: tt.requiredEnergy,
				Category:       tt.category,
			}
			result := analyzer.getOptimalTimeSlot(task)
			if result != tt.expected {
				t.Errorf("getOptimalTimeSlot() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestSortTasksByPriority(t *testing.T) {
	tasks := []PlannedTask{
		{Task: &taskwarrior.Task{Description: "Flexible task"}, Category: CategoryFlexible, Urgency: 5.0},
		{Task: &taskwarrior.Task{Description: "Critical task"}, Category: CategoryCritical, Urgency: 25.0},
		{Task: &taskwarrior.Task{Description: "Important task"}, Category: CategoryImportant, Urgency: 15.0},
		{Task: &taskwarrior.Task{Description: "Due critical task"}, Category: CategoryCritical, IsDue: true, Urgency: 20.0},
	}

	SortTasksByPriority(tasks)

	// Should be sorted: Critical (due) > Critical (not due) > Important > Flexible
	expectedOrder := []string{
		"Due critical task",
		"Critical task", 
		"Important task",
		"Flexible task",
	}

	for i, expected := range expectedOrder {
		if tasks[i].Description != expected {
			t.Errorf("Task %d: expected %s, got %s", i, expected, tasks[i].Description)
		}
	}
}

func TestCalculateWorkloadSummary(t *testing.T) {
	tasks := []PlannedTask{
		{EstimatedHours: 2.0, Category: CategoryCritical, RequiredEnergy: EnergyHigh},
		{EstimatedHours: 1.5, Category: CategoryImportant, RequiredEnergy: EnergyMedium},
		{EstimatedHours: 3.0, Category: CategoryFlexible, RequiredEnergy: EnergyLow},
	}

	totalHours, breakdown, energyBreakdown := CalculateWorkloadSummary(tasks)

	expectedTotal := 6.5
	if totalHours != expectedTotal {
		t.Errorf("Expected total hours %f, got %f", expectedTotal, totalHours)
	}

	expectedBreakdown := map[TaskCategory]float64{
		CategoryCritical:  2.0,
		CategoryImportant: 1.5,
		CategoryFlexible:  3.0,
	}

	for category, expected := range expectedBreakdown {
		if breakdown[category] != expected {
			t.Errorf("Category %v: expected %f hours, got %f", category, expected, breakdown[category])
		}
	}

	expectedEnergyBreakdown := map[EnergyLevel]float64{
		EnergyHigh:   2.0,
		EnergyMedium: 1.5,
		EnergyLow:    3.0,
	}

	for energy, expected := range expectedEnergyBreakdown {
		if energyBreakdown[energy] != expected {
			t.Errorf("Energy %v: expected %f hours, got %f", energy, expected, energyBreakdown[energy])
		}
	}
}

