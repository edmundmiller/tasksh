package daily

import (
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/planning/shared"
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

	// Mock some available tasks
	session.AvailableTasks = []shared.PlannedTask{
		{Task: &mockTask("task1"), EstimatedHours: 2.0},
		{Task: &mockTask("task2"), EstimatedHours: 1.5},
		{Task: &mockTask("task3"), EstimatedHours: 3.0},
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
		{Task: &mockTask("task1"), EstimatedHours: 2.0, RequiredEnergy: shared.EnergyHigh},
		{Task: &mockTask("task2"), EstimatedHours: 1.5, RequiredEnergy: shared.EnergyMedium},
		{Task: &mockTask("task3"), EstimatedHours: 3.0, RequiredEnergy: shared.EnergyLow},
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

	if total != int(StepCompleted) {
		t.Errorf("Expected total steps %d, got %d", int(StepCompleted), total)
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

// Mock helper functions

type mockTaskwarriorTask struct {
	description string
	uuid        string
}

func (m *mockTaskwarriorTask) GetDescription() string { return m.description }
func (m *mockTaskwarriorTask) GetUUID() string        { return m.uuid }

func mockTask(description string) *mockTaskwarriorTask {
	return &mockTaskwarriorTask{
		description: description,
		uuid:        "mock-uuid-" + description,
	}
}

// Note: This is a simplified mock that doesn't fully implement taskwarrior.Task interface
// In a real implementation, you'd want a more complete mock or use the actual Task struct