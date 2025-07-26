package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// TaskSuggestion represents an AI suggestion for a task
type TaskSuggestion struct {
	Type        string `json:"type"`        // "due_date", "priority", "estimate", "tag", "project"
	CurrentValue string `json:"current"`     // Current value
	SuggestedValue string `json:"suggested"` // Suggested value
	Reason      string `json:"reason"`      // Explanation for the suggestion
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
}

// TaskAnalysis represents the complete AI analysis of a task
type TaskAnalysis struct {
	TaskUUID    string           `json:"task_uuid"`
	Summary     string           `json:"summary"`
	Suggestions []TaskSuggestion `json:"suggestions"`
	TimeEstimate struct {
		Hours  float64 `json:"hours"`
		Reason string  `json:"reason"`
	} `json:"time_estimate"`
}

// AIAnalyzer handles AI-powered task analysis
type AIAnalyzer struct {
	timeDB *TimeDB
}

// NewAIAnalyzer creates a new AI analyzer
func NewAIAnalyzer(timeDB *TimeDB) *AIAnalyzer {
	return &AIAnalyzer{timeDB: timeDB}
}

// checkModsAvailable checks if the mods command is available
func checkModsAvailable() error {
	cmd := exec.Command("mods", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mods command not found: %w", err)
	}
	return nil
}

// AnalyzeTask performs AI analysis of a task using mods
func (ai *AIAnalyzer) AnalyzeTask(task *Task) (*TaskAnalysis, error) {
	if err := checkModsAvailable(); err != nil {
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
func (ai *AIAnalyzer) buildAnalysisPrompt(task *Task, estimate float64, estimateReason string, similar []TimeEntry) string {
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
	
	if len(similar) > 0 {
		prompt.WriteString("\n## Similar Completed Tasks\n")
		for i, entry := range similar {
			prompt.WriteString(fmt.Sprintf("%d. \"%s\" - %s hours (%s)\n", 
				i+1, entry.Description, formatHours(entry.ActualHours), entry.CompletedAt.Format("2006-01-02")))
		}
	}
	
	// Current date context
	prompt.WriteString(fmt.Sprintf("\n## Context\n"))
	prompt.WriteString(fmt.Sprintf("Today is %s\n", time.Now().Format("Monday, January 2, 2006")))
	
	// Request specific format
	prompt.WriteString("\n## Response Format\n")
	prompt.WriteString("Please respond with a JSON object containing:\n")
	prompt.WriteString("```json\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"summary\": \"Brief 1-2 sentence analysis\",\n")
	prompt.WriteString("  \"suggestions\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"type\": \"due_date|priority|estimate|tag|project\",\n")
	prompt.WriteString("      \"current\": \"current value\",\n")
	prompt.WriteString("      \"suggested\": \"suggested value\",\n")
	prompt.WriteString("      \"reason\": \"explanation\",\n")
	prompt.WriteString("      \"confidence\": 0.8\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ],\n")
	prompt.WriteString("  \"time_estimate\": {\n")
	prompt.WriteString("    \"hours\": 2.5,\n")
	prompt.WriteString("    \"reason\": \"Based on similar tasks\"\n")
	prompt.WriteString("  }\n")
	prompt.WriteString("}\n")
	prompt.WriteString("```\n\n")
	
	prompt.WriteString("Focus on practical improvements like:\n")
	prompt.WriteString("- Due date adjustments (too soon/far, missing deadlines)\n")
	prompt.WriteString("- Priority optimization (conflicts with other tasks)\n")
	prompt.WriteString("- Time estimates based on historical data\n")
	prompt.WriteString("- Better project/tag organization\n")
	prompt.WriteString("- Detecting missing dependencies or prerequisites\n\n")
	
	prompt.WriteString("Only suggest changes that would meaningfully improve task management. ")
	prompt.WriteString("If the task looks well-organized, say so in the summary and provide minimal suggestions.")
	
	return prompt.String()
}

// parseAnalysisResponse parses the JSON response from mods
func (ai *AIAnalyzer) parseAnalysisResponse(taskUUID, response string) (*TaskAnalysis, error) {
	// Extract JSON from the response (mods might include extra text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON found in response")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var analysis TaskAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	analysis.TaskUUID = taskUUID
	return &analysis, nil
}

// FormatAnalysis returns a human-readable format of the analysis
func (ai *AIAnalyzer) FormatAnalysis(analysis *TaskAnalysis) string {
	var output strings.Builder
	
	output.WriteString("🤖 AI Analysis\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	if analysis.Summary != "" {
		output.WriteString(fmt.Sprintf("**Summary**: %s\n\n", analysis.Summary))
	}
	
	if len(analysis.Suggestions) > 0 {
		output.WriteString("**Suggestions**:\n")
		for i, suggestion := range analysis.Suggestions {
			confidence := fmt.Sprintf("%.0f%%", suggestion.Confidence*100)
			output.WriteString(fmt.Sprintf("%d. **%s** [%s confidence]\n", i+1, strings.Title(strings.Replace(suggestion.Type, "_", " ", -1)), confidence))
			
			if suggestion.CurrentValue != "" {
				output.WriteString(fmt.Sprintf("   Current: %s → Suggested: %s\n", suggestion.CurrentValue, suggestion.SuggestedValue))
			} else {
				output.WriteString(fmt.Sprintf("   Suggested: %s\n", suggestion.SuggestedValue))
			}
			output.WriteString(fmt.Sprintf("   Reason: %s\n\n", suggestion.Reason))
		}
	}
	
	if analysis.TimeEstimate.Hours > 0 {
		output.WriteString(fmt.Sprintf("**Time Estimate**: %.1f hours\n", analysis.TimeEstimate.Hours))
		if analysis.TimeEstimate.Reason != "" {
			output.WriteString(fmt.Sprintf("Rationale: %s\n", analysis.TimeEstimate.Reason))
		}
	}
	
	return output.String()
}

// GetModificationSuggestions returns taskwarrior-compatible modification strings
func (ai *AIAnalyzer) GetModificationSuggestions(analysis *TaskAnalysis) []ModificationOption {
	var options []ModificationOption
	
	for _, suggestion := range analysis.Suggestions {
		var value, description string
		
		switch suggestion.Type {
		case "due_date":
			value = "due:" + suggestion.SuggestedValue
			description = fmt.Sprintf("Set due date to %s", suggestion.SuggestedValue)
		case "priority":
			value = "priority:" + suggestion.SuggestedValue
			description = fmt.Sprintf("Set priority to %s", suggestion.SuggestedValue)
		case "project":
			value = "project:" + suggestion.SuggestedValue
			description = fmt.Sprintf("Move to project %s", suggestion.SuggestedValue)
		case "tag":
			if strings.HasPrefix(suggestion.SuggestedValue, "+") {
				value = suggestion.SuggestedValue
				description = fmt.Sprintf("Add tag %s", suggestion.SuggestedValue[1:])
			} else {
				value = "+" + suggestion.SuggestedValue
				description = fmt.Sprintf("Add tag %s", suggestion.SuggestedValue)
			}
		case "estimate":
			// This could be a custom UDA for time estimates
			value = fmt.Sprintf("estimate:%.1fh", analysis.TimeEstimate.Hours)
			description = fmt.Sprintf("Set time estimate to %.1f hours", analysis.TimeEstimate.Hours)
		}
		
		if value != "" {
			options = append(options, ModificationOption{
				Name:        fmt.Sprintf("[AI] %s", description),
				Value:       value,
				Description: suggestion.Reason,
			})
		}
	}
	
	return options
}

// Helper functions
func getValueOrEmpty(value string) string {
	if value == "" {
		return "(none)"
	}
	return value
}

func formatHours(hours float64) string {
	if hours < 1 {
		return fmt.Sprintf("%.0f min", hours*60)
	}
	return fmt.Sprintf("%.1f", hours)
}