package timewarrior

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Entry represents a timewarrior time entry
type Entry struct {
	ID    int        `json:"id"`
	Start TimeWarriorTime `json:"start"`
	End   TimeWarriorTime `json:"end"`
	Tags  []string   `json:"tags"`
}

// TimeWarriorTime is a custom time type that can parse timewarrior's format
type TimeWarriorTime struct {
	time.Time
}

// UnmarshalJSON handles the timewarrior date format (20060102T150405Z)
func (t *TimeWarriorTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "" || str == "null" {
		t.Time = time.Time{}
		return nil
	}
	
	// Try parsing as timewarrior format first (basic ISO 8601)
	parsedTime, err := time.Parse("20060102T150405Z", str)
	if err != nil {
		// Try standard ISO 8601 format
		parsedTime, err = time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("failed to parse time %s: %w", str, err)
		}
	}
	
	t.Time = parsedTime
	return nil
}

// Client provides methods to interact with timewarrior
type Client struct {
	command string
}

// NewClient creates a new timewarrior client
func NewClient() *Client {
	return &Client{
		command: "timew",
	}
}

// Export exports time entries for a given time range
func (c *Client) Export(args ...string) ([]Entry, error) {
	cmdArgs := append([]string{"export"}, args...)
	cmd := exec.Command(c.command, cmdArgs...)
	
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("timewarrior export failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run timewarrior: %w", err)
	}
	
	// Handle empty output
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "[]" {
		return []Entry{}, nil
	}
	
	var entries []Entry
	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse timewarrior output: %w", err)
	}
	
	return entries, nil
}

// GetEntriesForTask returns all time entries for a specific task UUID
func (c *Client) GetEntriesForTask(taskUUID string) ([]Entry, error) {
	// Export entries with the task UUID tag
	taskTag := fmt.Sprintf("task_%s", taskUUID)
	entries, err := c.Export(taskTag)
	if err != nil {
		return nil, err
	}
	
	// Filter entries that have the task tag
	var taskEntries []Entry
	for _, entry := range entries {
		for _, tag := range entry.Tags {
			if tag == taskTag {
				taskEntries = append(taskEntries, entry)
				break
			}
		}
	}
	
	return taskEntries, nil
}

// GetEntriesForProject returns all time entries for a specific project
func (c *Client) GetEntriesForProject(projectName string) ([]Entry, error) {
	// Export entries with the project tag
	projectTag := fmt.Sprintf("project_%s", strings.ReplaceAll(projectName, " ", "_"))
	entries, err := c.Export(projectTag)
	if err != nil {
		return nil, err
	}
	
	// Filter entries that have the project tag
	var projectEntries []Entry
	for _, entry := range entries {
		for _, tag := range entry.Tags {
			if tag == projectTag {
				projectEntries = append(projectEntries, entry)
				break
			}
		}
	}
	
	return projectEntries, nil
}

// GetEntriesForDateRange returns all time entries within a date range
func (c *Client) GetEntriesForDateRange(start, end time.Time) ([]Entry, error) {
	// Format dates for timewarrior
	startStr := start.Format("2006-01-02T15:04:05")
	endStr := end.Format("2006-01-02T15:04:05")
	
	return c.Export(fmt.Sprintf("%s", startStr), "-", fmt.Sprintf("%s", endStr))
}

// CalculateTotalHours calculates total hours from a set of entries
func CalculateTotalHours(entries []Entry) float64 {
	totalDuration := time.Duration(0)
	
	for _, entry := range entries {
		if !entry.End.IsZero() {
			duration := entry.End.Sub(entry.Start.Time)
			totalDuration += duration
		}
	}
	
	return totalDuration.Hours()
}

// GetTaskDescription extracts the task description from tags
func GetTaskDescription(entry Entry) string {
	for _, tag := range entry.Tags {
		// Skip special tags
		if strings.HasPrefix(tag, "task_") || strings.HasPrefix(tag, "project_") {
			continue
		}
		// Return the first non-special tag as description
		return strings.ReplaceAll(tag, "_", " ")
	}
	return ""
}

// GetTaskUUID extracts the task UUID from tags
func GetTaskUUID(entry Entry) string {
	for _, tag := range entry.Tags {
		if strings.HasPrefix(tag, "task_") {
			return strings.TrimPrefix(tag, "task_")
		}
	}
	return ""
}

// GetProjectName extracts the project name from tags
func GetProjectName(entry Entry) string {
	for _, tag := range entry.Tags {
		if strings.HasPrefix(tag, "project_") {
			projectName := strings.TrimPrefix(tag, "project_")
			return strings.ReplaceAll(projectName, "_", " ")
		}
	}
	return ""
}