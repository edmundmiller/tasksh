package weekly

import (
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/planning/shared"
)

func TestNewWeeklyPlanningSession(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC) // Monday
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create weekly planning session: %v", err)
	}
	defer session.Close()

	if session == nil {
		t.Fatal("Session should not be nil")
	}

	if session.CurrentStep != StepWeeklyReflection {
		t.Errorf("Expected current step to be StepWeeklyReflection, got %v", session.CurrentStep)
	}

	if !session.WeekStart.Equal(weekStart) {
		t.Errorf("Expected week start %v, got %v", weekStart, session.WeekStart)
	}

	expectedWeekEnd := weekStart.AddDate(0, 0, 6) // Sunday
	if !session.WeekEnd.Equal(expectedWeekEnd) {
		t.Errorf("Expected week end %v, got %v", expectedWeekEnd, session.WeekEnd)
	}
}

func TestWeeklyWorkflowStepProgression(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test step progression
	steps := []WorkflowStep{
		StepWeeklyReflection,
		StepObjectiveSetting,
		StepJournaling,
		StepWorkStreamPlanning,
		StepWeeklySummary,
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

func TestWeeklyReflectionData(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no reflection
	if session.WeeklyReflection != nil {
		t.Error("Expected no initial reflection data")
	}

	// Set reflection data
	reflection := &WeeklyReflectionData{
		PreviousWeekStart:    weekStart.AddDate(0, 0, -7),
		PreviousWeekEnd:      weekStart.AddDate(0, 0, -1),
		KeyAccomplishments:   []string{"Completed project milestone"},
		LessonsLearned:      []string{"Better time estimation needed"},
		OverallSatisfaction: 4,
	}

	session.SetWeeklyReflection(reflection)

	if session.WeeklyReflection == nil {
		t.Error("Expected reflection data to be set")
	}

	if len(session.WeeklyReflection.KeyAccomplishments) != 1 {
		t.Errorf("Expected 1 accomplishment, got %d", len(session.WeeklyReflection.KeyAccomplishments))
	}
}

func TestObjectiveManagement(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no objectives
	if len(session.Objectives) != 0 {
		t.Error("Expected no initial objectives")
	}

	// Add objective
	obj := session.AddObjective("Complete project milestone", "Finish the core features", 1)
	if obj == nil {
		t.Fatal("Expected objective to be returned")
	}

	if len(session.Objectives) != 1 {
		t.Errorf("Expected 1 objective, got %d", len(session.Objectives))
	}

	if session.Objectives[0].Title != "Complete project milestone" {
		t.Errorf("Expected title 'Complete project milestone', got '%s'", session.Objectives[0].Title)
	}

	// Add another objective
	session.AddObjective("Improve team processes", "Streamline communication", 2)
	if len(session.Objectives) != 2 {
		t.Errorf("Expected 2 objectives, got %d", len(session.Objectives))
	}

	// Update objective
	err = session.UpdateObjective(0, "Updated milestone", "Updated description")
	if err != nil {
		t.Errorf("Failed to update objective: %v", err)
	}

	if session.Objectives[0].Title != "Updated milestone" {
		t.Errorf("Expected updated title, got '%s'", session.Objectives[0].Title)
	}

	// Remove objective
	err = session.RemoveObjective(0)
	if err != nil {
		t.Errorf("Failed to remove objective: %v", err)
	}

	if len(session.Objectives) != 1 {
		t.Errorf("Expected 1 objective after removal, got %d", len(session.Objectives))
	}

	// Test invalid index operations
	err = session.UpdateObjective(10, "invalid", "invalid")
	if err == nil {
		t.Error("Expected error for invalid index")
	}

	err = session.RemoveObjective(10)
	if err == nil {
		t.Error("Expected error for invalid index")
	}
}

func TestJournalEntry(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no journal entry
	if session.JournalEntry != "" {
		t.Error("Expected no initial journal entry")
	}

	// Set journal entry
	entry := "This week I want to focus on strategic initiatives and team building."
	session.SetJournalEntry(entry)

	if session.JournalEntry != entry {
		t.Errorf("Expected journal entry '%s', got '%s'", entry, session.JournalEntry)
	}

	// Test trimming whitespace
	session.SetJournalEntry("  whitespace test  ")
	if session.JournalEntry != "whitespace test" {
		t.Errorf("Expected trimmed entry, got '%s'", session.JournalEntry)
	}
}

func TestWorkStreamManagement(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Initially no work streams
	if len(session.WorkStreams) != 0 {
		t.Error("Expected no initial work streams")
	}

	// Create work stream
	ws := session.CreateWorkStream("Development Focus", "Core feature development", []string{"obj1"}, 20.0)
	if ws == nil {
		t.Fatal("Expected work stream to be returned")
	}

	if len(session.WorkStreams) != 1 {
		t.Errorf("Expected 1 work stream, got %d", len(session.WorkStreams))
	}

	if session.WorkStreams[0].Title != "Development Focus" {
		t.Errorf("Expected title 'Development Focus', got '%s'", session.WorkStreams[0].Title)
	}

	if session.WorkStreams[0].EstimatedHours != 20.0 {
		t.Errorf("Expected 20.0 hours, got %f", session.WorkStreams[0].EstimatedHours)
	}

	// Test adding tasks to work stream (with mock tasks)
	session.AvailableTasks = []shared.PlannedTask{
		{EstimatedHours: 2.0},
		{EstimatedHours: 1.5},
	}

	err = session.AddTaskToWorkStream(0, 0)
	if err != nil {
		t.Errorf("Failed to add task to work stream: %v", err)
	}

	if len(session.WorkStreams[0].Tasks) != 1 {
		t.Errorf("Expected 1 task in work stream, got %d", len(session.WorkStreams[0].Tasks))
	}

	// Test invalid indices
	err = session.AddTaskToWorkStream(10, 0)
	if err == nil {
		t.Error("Expected error for invalid work stream index")
	}

	err = session.AddTaskToWorkStream(0, 10)
	if err == nil {
		t.Error("Expected error for invalid task index")
	}
}

func TestGetObjectiveSuggestions(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	suggestions := session.GetObjectiveSuggestions()
	
	if len(suggestions) == 0 {
		t.Error("Expected at least some objective suggestions")
	}

	// Test with tasks that should generate project-based suggestions
	session.AvailableTasks = []shared.PlannedTask{
		{Task: &mockTask("task1", "ProjectA")},
		{Task: &mockTask("task2", "ProjectA")},
		{Task: &mockTask("task3", "ProjectA")},
		{Task: &mockTask("task4", "ProjectB")},
	}

	suggestions = session.GetObjectiveSuggestions()
	
	// Should include project-based suggestion for ProjectA (3+ tasks)
	found := false
	for _, suggestion := range suggestions {
		if suggestion == "Make significant progress on ProjectA project" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected project-based suggestion for ProjectA")
	}
}

func TestWeeklySummary(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Add some content to summarize
	session.AddObjective("Complete milestone", "Finish core features", 1)
	session.SetJournalEntry("Focus on strategic initiatives this week")
	session.CreateWorkStream("Development", "Core development work", []string{}, 15.0)

	summary := session.GetWeeklySummary()
	
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Check that summary contains key elements
	if !containsString(summary, "Complete milestone") {
		t.Error("Expected summary to contain objective title")
	}

	if !containsString(summary, "Development") {
		t.Error("Expected summary to contain work stream title")
	}

	if !containsString(summary, "15.0") {
		t.Error("Expected summary to contain work stream hours")
	}
}

func TestGetReflectionPrompts(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	prompts := session.GetReflectionPrompts()
	
	if len(prompts) == 0 {
		t.Error("Expected at least some reflection prompts")
	}

	// Check that prompts contain date information
	found := false
	for _, prompt := range prompts {
		if containsString(prompt, "Dec 4") && containsString(prompt, "Dec 10") {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected reflection prompts to contain previous week dates")
	}
}

func TestGetJournalingPrompts(t *testing.T) {
	weekStart := time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC)
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	prompts := session.GetJournalingPrompts()
	
	if len(prompts) == 0 {
		t.Error("Expected at least some journaling prompts")
	}

	// Check for strategic-focused prompts
	expectedPrompts := []string{
		"most important",
		"objectives",
		"obstacles",
		"successful",
	}

	for _, expected := range expectedPrompts {
		found := false
		for _, prompt := range prompts {
			if containsString(prompt, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find prompt containing '%s'", expected)
		}
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock task for testing
type mockTask struct {
	description string
	project     string
	uuid        string
}

func (m *mockTask) GetDescription() string { return m.description }
func (m *mockTask) GetProject() string     { return m.project }  
func (m *mockTask) GetUUID() string        { return m.uuid }

func mockTaskWithProject(description, project string) *mockTask {
	return &mockTask{
		description: description,
		project:     project,
		uuid:        "mock-uuid-" + description,
	}
}