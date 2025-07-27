package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

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

// AnalyzeTask performs AI analysis of a task using OpenAI API
func (ai *Analyzer) AnalyzeTask(task *taskwarrior.Task) (*TaskAnalysis, error) {
	if err := CheckOpenAIAvailable(); err != nil {
		return nil, err
	}

	// Get historical context
	estimate, estimateReason, _ := ai.timeDB.EstimateTimeForTask(task)
	similar, _ := ai.timeDB.GetSimilarTasks(task, 3)

	// Build the prompt
	prompt := ai.buildAnalysisPrompt(task, estimate, estimateReason, similar)

	// Get API key
	apiKey := ai.getOpenAIAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not available")
	}

	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Call OpenAI API
	ctx := context.Background()
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI API")
	}

	// Parse the response
	return ai.parseAnalysisResponse(task.UUID, resp.Choices[0].Message.Content)
}

// GenerateCommands uses AI to convert a natural language prompt into taskwarrior/timewarrior commands
func (ai *Analyzer) GenerateCommands(task *taskwarrior.Task, prompt string) ([]string, string, error) {
	if err := CheckOpenAIAvailable(); err != nil {
		return nil, "", err
	}

	// Get API key
	apiKey := ai.getOpenAIAPIKey()
	if apiKey == "" {
		return nil, "", fmt.Errorf("OpenAI API key not available")
	}

	// Build the agent prompt
	agentPrompt := ai.buildAgentPrompt(task, prompt)

	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Call OpenAI API
	ctx := context.Background()
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(agentPrompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return nil, "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, "", fmt.Errorf("no response from OpenAI API")
	}

	// Parse the commands response
	return ai.parseCommandsResponse(resp.Choices[0].Message.Content)
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

// getOpenAIAPIKey retrieves the OpenAI API key from environment or command
func (ai *Analyzer) getOpenAIAPIKey() string {
	// First try direct environment variable
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return apiKey
	}
	
	// Try the 1Password CLI command as mentioned in user instructions  
	if cmdStr := os.Getenv("OPENAI_API_KEY_CMD"); cmdStr != "" {
		// For the example command: $(op read "op://Private/api.openai.com/apikey")
		// We'll execute the command and return its output
		cmd := exec.Command("sh", "-c", cmdStr)
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	
	// Try the specific 1Password command mentioned by the user
	cmd := exec.Command("op", "read", "op://Private/api.openai.com/apikey")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	
	return ""
}

// buildAgentPrompt creates a prompt for the command generation agent
func (ai *Analyzer) buildAgentPrompt(task *taskwarrior.Task, userPrompt string) string {
	var prompt strings.Builder
	
	prompt.WriteString("# Taskwarrior/Timewarrior Command Agent\n\n")
	prompt.WriteString("You are an expert assistant that converts natural language requests into taskwarrior and timewarrior commands.\n\n")
	
	// Current task context
	prompt.WriteString("## Current Task\n")
	prompt.WriteString(fmt.Sprintf("- **UUID**: %s\n", task.UUID))
	prompt.WriteString(fmt.Sprintf("- **Description**: %s\n", task.Description))
	prompt.WriteString(fmt.Sprintf("- **Project**: %s\n", getValueOrEmpty(task.Project)))
	prompt.WriteString(fmt.Sprintf("- **Priority**: %s\n", getValueOrEmpty(task.Priority)))
	prompt.WriteString(fmt.Sprintf("- **Status**: %s\n", task.Status))
	if task.Due != "" {
		prompt.WriteString(fmt.Sprintf("- **Due**: %s\n", task.Due))
	}
	prompt.WriteString("\n")
	
	// User request
	prompt.WriteString("## User Request\n")
	prompt.WriteString(fmt.Sprintf("\"%s\"\n\n", userPrompt))
	
	// Instructions
	prompt.WriteString("## Instructions\n")
	prompt.WriteString("Convert the user's request into appropriate taskwarrior and/or timewarrior commands.\n\n")
	
	prompt.WriteString("**Available Commands:**\n")
	prompt.WriteString("- `task <uuid> done` - Mark task as completed\n")
	prompt.WriteString("- `task <uuid> modify <changes>` - Modify task attributes\n")
	prompt.WriteString("- `task <uuid> start` - Start tracking time for this task\n")
	prompt.WriteString("- `task <uuid> stop` - Stop tracking time for this task\n")
	prompt.WriteString("- `task <uuid> delete` - Delete the task\n")
	prompt.WriteString("- `timew start <description>` - Start time tracking\n")
	prompt.WriteString("- `timew stop` - Stop time tracking\n")
	prompt.WriteString("- `timew summary` - Show time summary\n")
	prompt.WriteString("- `timew track <start> - <end> <description>` - Log time retroactively\n\n")
	
	prompt.WriteString("**Response Format:**\n")
	prompt.WriteString("Provide your response in this JSON format:\n\n")
	prompt.WriteString("```json\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"commands\": [\n")
	prompt.WriteString("    \"task <uuid> done\",\n")
	prompt.WriteString("    \"timew track 2h ago - now 'Working on task'\"\n")
	prompt.WriteString("  ],\n")
	prompt.WriteString("  \"explanation\": \"This will mark the task as complete and log 2 hours of work time.\",\n")
	prompt.WriteString("  \"safety_note\": \"These commands will modify your task and time data. Review carefully before executing.\"\n")
	prompt.WriteString("}\n")
	prompt.WriteString("```\n\n")
	
	prompt.WriteString("**Important Notes:**\n")
	prompt.WriteString("- Use the exact task UUID provided\n")
	prompt.WriteString("- Be conservative with destructive actions\n")
	prompt.WriteString("- If the request is unclear, ask for clarification in the explanation\n")
	prompt.WriteString("- Always provide clear explanations of what each command does\n")
	
	return prompt.String()
}

// parseCommandsResponse parses the AI response for command generation
func (ai *Analyzer) parseCommandsResponse(response string) ([]string, string, error) {
	// Extract JSON from response (it might be wrapped in markdown)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, "", fmt.Errorf("no JSON found in response")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var result struct {
		Commands    []string `json:"commands"`
		Explanation string   `json:"explanation"`
		SafetyNote  string   `json:"safety_note"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, "", fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	// Build preview text
	preview := fmt.Sprintf("**Explanation:** %s\n\n**Commands to execute:**\n", result.Explanation)
	for i, cmd := range result.Commands {
		preview += fmt.Sprintf("%d. `%s`\n", i+1, cmd)
	}
	if result.SafetyNote != "" {
		preview += fmt.Sprintf("\n**⚠️  %s**", result.SafetyNote)
	}
	
	return result.Commands, preview, nil
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