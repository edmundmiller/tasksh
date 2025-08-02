package daily

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/planning/shared"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestNewPlanningModel(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	
	if model == nil {
		t.Fatal("Expected non-nil model")
	}
	
	if model.session != session {
		t.Error("Expected model to reference the session")
	}
	
	if model.width != 80 {
		t.Errorf("Expected default width 80, got %d", model.width)
	}
	
	if model.height != 24 {
		t.Errorf("Expected default height 24, got %d", model.height)
	}
}

func TestModelInit(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	cmd := model.Init()
	
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
	
	// Check that background loading starts if on reflection step
	if session.CurrentStep == StepReflection {
		cmds := tea.Batch(cmd)
		if cmds == nil {
			t.Error("Expected batch command for reflection step")
		}
	}
}

func TestModelView(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	
	// Test quitting state
	model.quitting = true
	view := model.View()
	if !strings.Contains(view, "Daily planning session ended") {
		t.Error("Expected quit message when quitting")
	}
	
	// Test normal view
	model.quitting = false
	model.width = 80
	model.height = 24
	view = model.View()
	
	// Should contain header
	if !strings.Contains(view, "Daily Planning") {
		t.Error("Expected view to contain planning header")
	}
	
	// Should show step info
	if !strings.Contains(view, "Step") {
		t.Error("Expected view to show step progress")
	}
}

func TestRenderSteps(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24

	// Test reflection step
	session.CurrentStep = StepReflection
	content := model.renderReflectionStep()
	if !strings.Contains(content, "Starting your daily planning ritual") {
		t.Error("Expected reflection step to contain intro text")
	}

	// Test task selection step
	session.CurrentStep = StepTaskSelection
	content = model.renderTaskSelectionStep()
	if !strings.Contains(content, "Select tasks for today") {
		t.Error("Expected task selection step header")
	}

	// Test workload assessment step
	session.CurrentStep = StepWorkloadAssessment
	session.SelectedTasks = []shared.PlannedTask{
		{
			Task: &taskwarrior.Task{
				UUID:        "test-1",
				Description: "Test task",
			},
			EstimatedHours: 2.0,
			Category:       shared.CategoryImportant,
		},
	}
	content = model.renderWorkloadAssessmentStep()
	if !strings.Contains(content, "Workload Assessment") {
		t.Error("Expected workload assessment header")
	}

	// Test finalization step
	session.CurrentStep = StepFinalization
	content = model.renderFinalizationStep()
	if !strings.Contains(content, "Finalize Your Plan") {
		t.Error("Expected finalization header")
	}

	// Test summary step
	session.CurrentStep = StepSummary
	content = model.renderSummaryStep()
	if !strings.Contains(content, "Daily Plan Summary") {
		t.Error("Expected summary header")
	}

	// Test completed step
	session.CurrentStep = StepCompleted
	content = model.renderCompletedStep()
	if !strings.Contains(content, "Daily Planning Complete") {
		t.Error("Expected completion header")
	}
}

func TestHandleKeyMessages(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24

	// Add some test tasks
	session.CurrentStep = StepTaskSelection
	session.AvailableTasks = []shared.PlannedTask{
		{
			Task: &taskwarrior.Task{
				UUID:        "test-1",
				Description: "Task 1",
			},
			EstimatedHours: 1.0,
		},
		{
			Task: &taskwarrior.Task{
				UUID:        "test-2",
				Description: "Task 2",
			},
			EstimatedHours: 2.0,
		},
	}

	// Test quit key
	quitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd := model.Update(quitMsg)
	planningModel := updatedModel.(*PlanningModel)
	if !planningModel.quitting {
		t.Error("Expected model to be in quitting state after 'q' key")
	}
	if cmd == nil {
		t.Error("Expected quit command")
	}

	// Reset and test navigation
	model.quitting = false
	model.selectedIndex = 0

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = model.Update(downMsg)
	planningModel = updatedModel.(*PlanningModel)
	if planningModel.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after down key, got %d", planningModel.selectedIndex)
	}

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = model.Update(upMsg)
	planningModel = updatedModel.(*PlanningModel)
	if planningModel.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after up key, got %d", planningModel.selectedIndex)
	}
}

func TestCapacityVisualization(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24

	// Set up tasks with different workloads
	session.CurrentStep = StepWorkloadAssessment
	session.SelectedTasks = []shared.PlannedTask{
		{
			Task: &taskwarrior.Task{
				UUID:        "test-1",
				Description: "Task 1",
			},
			EstimatedHours: 3.0,
			Category:       shared.CategoryCritical,
			RequiredEnergy: shared.EnergyHigh,
		},
		{
			Task: &taskwarrior.Task{
				UUID:        "test-2",
				Description: "Task 2",
			},
			EstimatedHours: 4.0,
			Category:       shared.CategoryImportant,
			RequiredEnergy: shared.EnergyMedium,
		},
	}

	// Calculate assessment
	session.CalculateWorkloadAssessment()
	
	content := model.renderWorkloadAssessmentStep()
	
	// Check for capacity bar
	if !strings.Contains(content, "Capacity:") {
		t.Error("Expected capacity visualization")
	}
	
	// Check for task breakdown
	if !strings.Contains(content, "TASK BREAKDOWN BY PRIORITY") {
		t.Error("Expected task breakdown section")
	}
	
	// Check for energy requirements
	if !strings.Contains(content, "ENERGY REQUIREMENTS") {
		t.Error("Expected energy requirements section")
	}
	
	// Check for recommendations
	if strings.Contains(content, "room for more tasks") {
		t.Error("Should not suggest room for more tasks when near capacity")
	}
}

func TestRenderHeader(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24

	header := model.renderHeader()
	
	// Check step display
	if !strings.Contains(header, "Step 1/5") {
		t.Error("Expected correct step numbering")
	}
	
	// Check step name
	if !strings.Contains(header, "Reflect on Yesterday") {
		t.Error("Expected step name in header")
	}
	
	// Check date
	if !strings.Contains(header, time.Now().Format("Monday, January 2, 2006")) {
		t.Error("Expected date in header")
	}
}

func TestLoadingState(t *testing.T) {
	session, err := NewDailyPlanningSession(time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24
	model.isLoading = true
	model.loadingMessage = "Loading tasks..."

	view := model.View()
	if !strings.Contains(view, model.loadingMessage) {
		t.Error("Expected loading message in view when isLoading is true")
	}
}