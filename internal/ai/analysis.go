package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// TaskSuggestion represents an AI suggestion for a task
type TaskSuggestion struct {
	Type           string  `json:"type"`        // "due_date", "priority", "estimate", "tag", "project"
	CurrentValue   string  `json:"current"`     // Current value
	SuggestedValue string  `json:"suggested"`   // Suggested value
	Reason         string  `json:"reason"`      // Explanation for the suggestion
	Confidence     float64 `json:"confidence"`  // 0.0 to 1.0
}

// TaskAnalysis represents the complete AI analysis of a task
type TaskAnalysis struct {
	TaskUUID     string           `json:"task_uuid"`
	Summary      string           `json:"summary"`
	Suggestions  []TaskSuggestion `json:"suggestions"`
	TimeEstimate struct {
		Hours  float64 `json:"hours"`
		Reason string  `json:"reason"`
	} `json:"time_estimate"`
}

// Analyzer handles AI-powered task analysis
type Analyzer struct {
	timeDB *timedb.TimeDB
}

// NewAnalyzer creates a new AI analyzer
func NewAnalyzer(timeDB *timedb.TimeDB) *Analyzer {
	return &Analyzer{timeDB: timeDB}
}

// AnalyzeTask performs AI analysis of a task using mods
func (ai *Analyzer) AnalyzeTask(task *taskwarrior.Task) (*TaskAnalysis, error) {
	if err := CheckModsAvailable(); err != nil {
		return nil, err
	}

	// Get historical context
	estimate, estimateReason, _ := ai.timeDB.EstimateTimeForTask(task)
	similar, _ := ai.timeDB.GetSimilarTasks(task, 3)

	// Build the prompt
	prompt := ai.buildAnalysisPrompt(task, estimate, estimateReason, similar)

	// Call mods
	cmd := exec.Command("mods", "--no-limit")
	cmd.Stdin = strings.NewReader(prompt)
	
	var output bytes.Buffer
	cmd.Stdout = &output
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mods execution failed: %w", err)
	}

	// Parse the response
	return ai.parseAnalysisResponse(task.UUID, output.String())
}

// buildAnalysisPrompt creates a structured prompt for task analysis
func (ai *Analyzer) buildAnalysisPrompt(task *taskwarrior.Task, estimate float64, estimateReason string, similar []timedb.TimeEntry) string {
	var prompt strings.Builder
	
	prompt.WriteString("# Task Analysis Request\n\n")
	prompt.WriteString("You are a task management expert helping to optimize a task. ")
	prompt.WriteString("Analyze the following task and provide specific, actionable suggestions.\n\n")
	
	// Current task details
	prompt.WriteString("## Current Task\n")
	prompt.WriteString(fmt.Sprintf("- **Description**: %s\n", task.Description))
	prompt.WriteString(fmt.Sprintf("- **Project**: %s\n", getValueOrEmpty(task.Project)))
	prompt.WriteString(fmt.Sprintf("- **Priority**: %s\n", getValueOrEmpty(task.Priority)))
	prompt.WriteString(fmt.Sprintf("- **Due Date**: %s\n", getValueOrEmpty(task.Due)))
	prompt.WriteString(fmt.Sprintf("- **Status**: %s\n", task.Status))
	
	// Historical context
	if estimate > 0 {
		prompt.WriteString(fmt.Sprintf("\n## Time Estimate\n"))
		prompt.WriteString(fmt.Sprintf("Historical estimate: %.1f hours (%s)\n", estimate, estimateReason))
	}
	
	// Similar tasks
	if len(similar) > 0 {
		prompt.WriteString("\n## Similar Completed Tasks\n")
		for i, entry := range similar {
			prompt.WriteString(fmt.Sprintf("%d. **%s** (project: %s, priority: %s)\n", 
				i+1, entry.Description, entry.Project, entry.Priority))
			prompt.WriteString(fmt.Sprintf("   - Estimated: %.1f hrs, Actual: %.1f hrs\n", 
				entry.EstimatedHours, entry.ActualHours))
		}
	}
	
	// Request specific format
	prompt.WriteString("\n## Analysis Request\n")
	prompt.WriteString("Please provide your analysis in the following JSON format:\n\n")
	prompt.WriteString("```json\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"summary\": \"Brief summary of the task and key observations\",\n")
	prompt.WriteString("  \"suggestions\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"type\": \"priority|due_date|project|tag|estimate\",\n")
	prompt.WriteString("      \"current\": \"current value or empty\",\n")
	prompt.WriteString("      \"suggested\": \"suggested value\",\n")
	prompt.WriteString("      \"reason\": \"explanation for suggestion\",\n")
	prompt.WriteString("      \"confidence\": 0.8\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ],\n")
	prompt.WriteString("  \"time_estimate\": {\n")
	prompt.WriteString("    \"hours\": 2.5,\n")
	prompt.WriteString("    \"reason\": \"explanation for time estimate\"\n")
	prompt.WriteString("  }\n")
	prompt.WriteString("}\n")
	prompt.WriteString("```\n\n")
	
	prompt.WriteString("Focus on practical improvements that will help with task completion and organization.")
	
	return prompt.String()
}

// parseAnalysisResponse parses the AI response into structured data
func (ai *Analyzer) parseAnalysisResponse(taskUUID, response string) (*TaskAnalysis, error) {
	// Extract JSON from response (it might be wrapped in markdown)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var analysis TaskAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	analysis.TaskUUID = taskUUID
	return &analysis, nil
}

// getValueOrEmpty returns the value or "none" if empty
func getValueOrEmpty(value string) string {
	if value == "" {
		return "none"
	}
	return value
}