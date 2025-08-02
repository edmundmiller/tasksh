package taskwarrior

import (
	"strings"
	"testing"
)

// TestExecuteTask tests the basic command execution
func TestExecuteTask(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Test a simple version command
	output, err := executeTask("version")
	if err != nil {
		t.Errorf("Failed to execute task version: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty version output")
	}
}

// TestAddTask tests task creation
func TestAddTask(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Test adding a simple task
	description := "Test task for unit testing"
	project := "project:testing"
	tags := []string{"+test", "+automated"}

	uuid, err := AddTask(description, project, tags...)
	if err != nil {
		t.Errorf("Failed to add task: %v", err)
	}

	// UUID might be empty in some configurations, but no error is success
	t.Logf("Created task with UUID: %s", uuid)

	// Clean up - delete the test task if we got a UUID
	if uuid != "" {
		DeleteTask(uuid)
	}
}

// TestGetTasksJSON tests JSON task retrieval
func TestGetTasksJSON(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Test getting tasks as JSON
	output, err := GetTasksJSON("status:pending", "limit:5")
	if err != nil {
		t.Errorf("Failed to get tasks JSON: %v", err)
	}

	// Output should be valid JSON array (starts with [ or is empty [])
	trimmed := strings.TrimSpace(output)
	if trimmed != "" && !strings.HasPrefix(trimmed, "[") {
		t.Error("Expected JSON array output")
	}
}

// TestCompleteTask tests task completion
func TestCompleteTask(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// First create a task to complete
	description := "Test task to complete"
	uuid, err := AddTask(description, "", "+test")
	if err != nil || uuid == "" {
		t.Skip("Could not create test task")
	}

	// Complete the task
	err = CompleteTask(uuid)
	if err != nil {
		t.Errorf("Failed to complete task: %v", err)
	}

	// Clean up - delete the completed task
	DeleteTask(uuid)
}

// TestModifyTask tests task modification
func TestModifyTask(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Create a test task
	description := "Test task to modify"
	uuid, err := AddTask(description, "", "+test")
	if err != nil || uuid == "" {
		t.Skip("Could not create test task")
	}

	// Modify the task
	err = ModifyTask(uuid, "priority:H")
	if err != nil {
		t.Errorf("Failed to modify task: %v", err)
	}

	// Clean up
	DeleteTask(uuid)
}

// TestRecurringTaskCreation tests creating a recurring task
func TestRecurringTaskCreation(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Create a recurring task
	description := "Test recurring task"
	uuid, err := AddTask(description, "project:test", "+recurring", "recur:daily", "due:tomorrow")
	if err != nil {
		t.Errorf("Failed to create recurring task: %v", err)
	}

	t.Logf("Created recurring task with UUID: %s", uuid)

	// Check if it was created as a recurring task
	output, err := GetTasksJSON("status:recurring", "description:Test recurring task")
	if err != nil {
		t.Errorf("Failed to check recurring task: %v", err)
	}

	if !strings.Contains(output, "Test recurring task") {
		t.Error("Recurring task not found in recurring tasks")
	}

	// Clean up - delete the recurring task
	if uuid != "" {
		DeleteTask(uuid)
	}
}

// TestBatchLoadTasks tests batch loading functionality
func TestBatchLoadTasks(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Create a few test tasks
	var uuids []string
	for i := 0; i < 3; i++ {
		description := "Batch test task"
		uuid, err := AddTask(description, "", "+batchtest")
		if err == nil && uuid != "" {
			uuids = append(uuids, uuid)
		}
	}

	if len(uuids) == 0 {
		t.Skip("Could not create test tasks")
	}

	// Batch load the tasks
	taskMap, err := BatchLoadTasks(uuids)
	if err != nil {
		t.Errorf("Failed to batch load tasks: %v", err)
	}

	// Verify we got all tasks
	if len(taskMap) != len(uuids) {
		t.Errorf("Expected %d tasks, got %d", len(uuids), len(taskMap))
	}

	// Clean up
	for _, uuid := range uuids {
		DeleteTask(uuid)
	}
}

// TestErrorHandling tests error handling for invalid operations
func TestErrorHandling(t *testing.T) {
	// Skip if taskwarrior is not available
	if err := CheckAvailable(); err != nil {
		t.Skip("Taskwarrior not available:", err)
	}

	// Test completing non-existent task
	err := CompleteTask("non-existent-uuid")
	if err == nil {
		t.Error("Expected error when completing non-existent task")
	}

	// Test modifying non-existent task
	err = ModifyTask("non-existent-uuid", "priority:H")
	if err == nil {
		t.Error("Expected error when modifying non-existent task")
	}

	// Test deleting non-existent task
	err = DeleteTask("non-existent-uuid")
	if err == nil {
		t.Error("Expected error when deleting non-existent task")
	}
}