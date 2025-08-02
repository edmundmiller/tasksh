package daily

import (
	"strings"
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/planning/shared"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestNewDailyPlanningSession(t *testing.T) {
	targetDate := time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC)
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create daily planning session: %v", err)
	}
	defer session.Close()

	if session == nil {
		t.Fatal("Session should not be nil")
	}

	if session.CurrentStep != StepReflection {
		t.Errorf("Expected current step to be StepReflection, got %v", session.CurrentStep)
	}

	if !session.Context.Date.Equal(targetDate) {
		t.Errorf("Expected target date %v, got %v", targetDate, session.Context.Date)
	}

	if session.Context.WorkdayHours != 8.0 {
		t.Errorf("Expected workday hours to be 8.0, got %f", session.Context.WorkdayHours)
	}
}

func TestWorkflowStepProgression(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test step progression
	steps := []WorkflowStep{
		StepReflection,
		StepTaskSelection,
		StepWorkloadAssessment,
		StepFinalization,
		StepSummary,
		StepCompleted,
	}

	for i, expectedStep := range steps {
		if session.CurrentStep != expectedStep {
			t.Errorf("Step %d: expected %v, got %v", i, expectedStep, session.CurrentStep)
		}

		if i < len(steps)-1 { // Don't advance from the last step
			err := session.NextStep()
			if err != nil {
				t.Errorf("Failed to advance to next step: %v", err)
			}
		}
	}

	// Test that we can't advance beyond completed
	err = session.NextStep()
	if err == nil {
		t.Error("Expected error when advancing beyond completed step")
	}
}

func TestWorkflowStepRegression(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Advance to task selection
	err = session.NextStep()
	if err != nil {
		t.Fatalf("Failed to advance step: %v", err)
	}

	// Test going back
	err = session.PreviousStep()
	if err != nil {
		t.Errorf("Failed to go back to previous step: %v", err)
	}

	if session.CurrentStep != StepReflection {
		t.Errorf("Expected to be back at StepReflection, got %v", session.CurrentStep)
	}

	// Test that we can't go back from the first step
	err = session.PreviousStep()
	if err == nil {
		t.Error("Expected error when going back from first step")
	}
}

func TestReflectionData(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no reflection
	if session.Reflection != nil {
		t.Error("Expected no initial reflection data")
	}

	// Set reflection data
	reflection := &shared.ReflectionData{
		Date:        targetDate.AddDate(0, 0, -1),
		EnergyLevel: shared.EnergyHigh,
		Accomplishments: []string{"Completed important project"},
	}

	session.SetReflectionData(reflection)

	if session.Reflection == nil {
		t.Error("Expected reflection data to be set")
	}

	if session.Context.EnergyLevel != shared.EnergyHigh {
		t.Errorf("Expected energy level to be updated to High, got %v", session.Context.EnergyLevel)
	}
}

func TestTaskSelection(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Clear any pre-selected tasks
	session.SelectedTasks = []shared.PlannedTask{}

	// Mock some available tasks
	// We need to use real Task structs, not mocks, since PlannedTask expects *taskwarrior.Task
	session.AvailableTasks = []shared.PlannedTask{
		{Task: &taskwarrior.Task{UUID: "task1", Description: "Task 1"}, EstimatedHours: 2.0},
		{Task: &taskwarrior.Task{UUID: "task2", Description: "Task 2"}, EstimatedHours: 1.5},
		{Task: &taskwarrior.Task{UUID: "task3", Description: "Task 3"}, EstimatedHours: 3.0},
	}

	// Test adding tasks to selection
	err = session.AddTaskToSelection(0)
	if err != nil {
		t.Errorf("Failed to add task to selection: %v", err)
	}

	if len(session.SelectedTasks) != 1 {
		t.Errorf("Expected 1 selected task, got %d", len(session.SelectedTasks))
	}

	// Test adding duplicate task
	err = session.AddTaskToSelection(0)
	if err == nil {
		t.Error("Expected error when adding duplicate task")
	}

	// Test removing task from selection
	err = session.RemoveTaskFromSelection(0)
	if err != nil {
		t.Errorf("Failed to remove task from selection: %v", err)
	}

	if len(session.SelectedTasks) != 0 {
		t.Errorf("Expected 0 selected tasks, got %d", len(session.SelectedTasks))
	}
}

func TestWorkloadAssessment(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Add some selected tasks
	session.SelectedTasks = []shared.PlannedTask{
		{Task: &taskwarrior.Task{UUID: "task1", Description: "Task 1"}, EstimatedHours: 2.0, RequiredEnergy: shared.EnergyHigh},
		{Task: &taskwarrior.Task{UUID: "task2", Description: "Task 2"}, EstimatedHours: 1.5, RequiredEnergy: shared.EnergyMedium},
		{Task: &taskwarrior.Task{UUID: "task3", Description: "Task 3"}, EstimatedHours: 3.0, RequiredEnergy: shared.EnergyLow},
	}

	assessment := session.CalculateWorkloadAssessment()

	if assessment == nil {
		t.Fatal("Expected workload assessment to be calculated")
	}

	if assessment.AvailableHours != 8.0 {
		t.Errorf("Expected 8.0 available hours, got %f", assessment.AvailableHours)
	}

	if assessment.FocusHours != 5.2 { // 8.0 * 0.65
		t.Errorf("Expected 5.2 focus hours, got %f", assessment.FocusHours)
	}

	// Test capacity warning
	warning := session.GetCapacityWarning()
	if warning == "" {
		t.Error("Expected capacity warning to be generated")
	}
}

func TestDailyFocus(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no focus
	if session.DailyFocus != "" {
		t.Error("Expected no initial daily focus")
	}

	// Set daily focus
	focus := "Complete project milestone"
	session.SetDailyFocus(focus)

	if session.DailyFocus != focus {
		t.Errorf("Expected daily focus '%s', got '%s'", focus, session.DailyFocus)
	}

	// Test trimming whitespace
	session.SetDailyFocus("  whitespace test  ")
	if session.DailyFocus != "whitespace test" {
		t.Errorf("Expected trimmed focus, got '%s'", session.DailyFocus)
	}
}

func TestGetStepProgress(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	current, total := session.GetStepProgress()
	
	if current != 0 { // StepReflection is 0
		t.Errorf("Expected current step 0, got %d", current)
	}

	// Total should be StepSummary + 1 (not StepCompleted)
	expectedTotal := int(StepSummary) + 1
	if total != expectedTotal {
		t.Errorf("Expected total steps %d, got %d", expectedTotal, total)
	}

	// Advance a step and test again
	session.NextStep()
	current, total = session.GetStepProgress()
	
	if current != 1 { // StepTaskSelection is 1
		t.Errorf("Expected current step 1, got %d", current)
	}
}

func TestIsCompleted(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially not completed
	if session.IsCompleted() {
		t.Error("Expected session to not be completed initially")
	}

	// Advance to completed state
	for !session.IsCompleted() {
		err := session.NextStep()
		if err != nil {
			break
		}
	}

	if !session.IsCompleted() {
		t.Error("Expected session to be completed")
	}

	if session.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

// Test background task loading
func TestBackgroundTaskLoading(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially tasks are not loaded
	if session.IsTasksLoaded() {
		t.Error("Expected tasks to not be loaded initially")
	}

	// Start background loading
	loadCmd := session.StartBackgroundTaskLoading()
	if loadCmd == nil {
		t.Error("Expected background loading command to be returned")
	}

	// Execute the loading function
	msg := loadCmd()
	if result, ok := msg.(BackgroundLoadResult); ok {
		// In a test environment without Taskwarrior, this might fail
		// but we're testing the mechanism works
		if result.Error != nil {
			t.Logf("Background loading error (expected in test): %v", result.Error)
		}
	} else {
		t.Error("Expected BackgroundLoadResult message type")
	}
}

// Test existing scheduled tasks detection
func TestCheckScheduledTasks(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// This will likely return empty in test environment
	// but we're testing the function exists and doesn't panic
	tasks, err := session.CheckScheduledTasks()
	if err != nil {
		// Expected in test environment without Taskwarrior
		t.Logf("CheckScheduledTasks error (expected): %v", err)
	} else {
		// If it somehow succeeds, verify the return type
		if tasks == nil {
			t.Error("Expected non-nil task slice")
		}
	}
}

// Test daily summary generation
func TestGetDailySummary(t *testing.T) {
	targetDate := time.Now()
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test with no tasks
	// Clear any pre-selected tasks
	session.SelectedTasks = []shared.PlannedTask{}
	
	summary := session.GetDailySummary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}
	if !strings.Contains(summary, "No tasks planned for today") {
		t.Error("Expected summary to indicate no tasks")
	}

	// Add some tasks
	session.SelectedTasks = []shared.PlannedTask{
		{
			Task: &taskwarrior.Task{
				UUID:        "test-1",
				Description: "Critical task",
			},
			EstimatedHours: 2.0,
			Category:       shared.CategoryCritical,
		},
		{
			Task: &taskwarrior.Task{
				UUID:        "test-2",
				Description: "Important task",
			},
			EstimatedHours: 1.5,
			Category:       shared.CategoryImportant,
		},
	}
	session.DailyFocus = "Test focus"

	summary = session.GetDailySummary()
	if !strings.Contains(summary, "Daily Plan for") {
		t.Error("Expected summary to contain date header")
	}
	if !strings.Contains(summary, "Test focus") {
		t.Error("Expected summary to contain daily focus")
	}
	if !strings.Contains(summary, "CRITICAL TASKS") {
		t.Error("Expected summary to contain critical tasks section")
	}
	if !strings.Contains(summary, "IMPORTANT TASKS") {
		t.Error("Expected summary to contain important tasks section")
	}
	if !strings.Contains(summary, "3.5 hours") {
		t.Error("Expected summary to contain total hours")
	}
}

