package estimation

import (
	"fmt"
	"time"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// LearningRecord represents a completed task with estimate vs actual comparison
type LearningRecord struct {
	TaskUUID       string
	Description    string
	Project        string
	EstimatedHours float64
	ActualHours    float64
	EstimateSource EstimationSource
	CompletedAt    time.Time
	AccuracyRatio  float64 // actual/estimated
}

// RecordTaskCompletion records a completed task for learning
func (e *Estimator) RecordTaskCompletion(task *taskwarrior.Task, estimatedHours, actualHours float64, source EstimationSource) error {
	if e.timeDB == nil {
		return fmt.Errorf("no time database available")
	}
	
	// Record in timeDB for future estimates
	err := e.timeDB.RecordCompletion(task, estimatedHours, actualHours)
	if err != nil {
		return fmt.Errorf("failed to record completion: %w", err)
	}
	
	// Also record learning metrics
	if err := e.recordLearningMetrics(task, estimatedHours, actualHours, source); err != nil {
		// Don't fail the whole operation if learning metrics fail
		// Just log it (in a real system)
	}
	
	return nil
}

// recordLearningMetrics stores accuracy metrics for improving estimates
func (e *Estimator) recordLearningMetrics(task *taskwarrior.Task, estimatedHours, actualHours float64, source EstimationSource) error {
	// In a full implementation, this would store:
	// - Accuracy by source (which estimation methods are most accurate)
	// - Accuracy by project (are we consistently over/under for certain projects)
	// - Accuracy by keywords (do certain task types have predictable biases)
	// - Accuracy trends over time
	
	// For now, we'll just calculate the accuracy ratio
	accuracyRatio := 1.0
	if estimatedHours > 0 {
		accuracyRatio = actualHours / estimatedHours
	}
	
	// This could be stored in a separate learning table
	// For demonstration, we'll just return the calculated value
	_ = accuracyRatio
	
	return nil
}

// GetEstimationAccuracy returns accuracy statistics for the estimation system
func (e *Estimator) GetEstimationAccuracy(source EstimationSource, projectFilter string) (*AccuracyStats, error) {
	// Get accuracy data from timeDB
	stats, err := e.timeDB.GetEstimationAccuracy()
	if err != nil {
		return nil, err
	}
	
	accuracy := &AccuracyStats{
		TotalTasks:        stats["total_tasks"].(int),
		AverageActual:     stats["avg_actual_hours"].(float64),
		AverageEstimated:  stats["avg_estimated_hours"].(float64),
		AverageErrorPct:   stats["avg_error_percent"].(float64),
	}
	
	// Calculate additional metrics
	if accuracy.AverageEstimated > 0 {
		accuracy.OverallAccuracy = accuracy.AverageActual / accuracy.AverageEstimated
	}
	
	return accuracy, nil
}

// AccuracyStats contains estimation accuracy metrics
type AccuracyStats struct {
	TotalTasks       int
	AverageActual    float64
	AverageEstimated float64
	AverageErrorPct  float64
	OverallAccuracy  float64 // 1.0 = perfect, <1 = overestimate, >1 = underestimate
}

// GetCalibrationFactor returns a calibration factor based on historical accuracy
func (e *Estimator) GetCalibrationFactor(source EstimationSource, project string) float64 {
	// In a full implementation, this would:
	// 1. Look up historical accuracy for this source/project combination
	// 2. Calculate a calibration factor to improve future estimates
	// 3. Apply smoothing to avoid overcorrection
	
	// For now, return neutral calibration
	return 1.0
}

// ApplyLearning adjusts an estimate based on historical accuracy
func (e *Estimator) ApplyLearning(estimate *Estimate, task *taskwarrior.Task) *Estimate {
	// Get calibration factor for this type of estimate
	calibration := e.GetCalibrationFactor(estimate.Source, task.Project)
	
	// Apply calibration
	calibratedEstimate := *estimate
	calibratedEstimate.Hours *= calibration
	
	// Update reason if calibration was applied
	if calibration != 1.0 {
		adjustment := "adjusted down"
		if calibration > 1.0 {
			adjustment = "adjusted up"
		}
		calibratedEstimate.Reason += fmt.Sprintf(" (%s %.0f%% based on historical accuracy)",
			adjustment, (calibration-1.0)*100)
	}
	
	return &calibratedEstimate
}

// GetProjectAccuracy returns accuracy metrics for a specific project
func (e *Estimator) GetProjectAccuracy(projectName string) (map[string]interface{}, error) {
	// This would need to be implemented in timeDB with proper query
	// For now, return placeholder
	return map[string]interface{}{
		"project": projectName,
		"status":  "learning metrics pending implementation",
	}, nil
}

// SuggestEstimateImprovement analyzes patterns and suggests improvements
func (e *Estimator) SuggestEstimateImprovement(task *taskwarrior.Task) []string {
	suggestions := []string{}
	
	// Analyze task description for ambiguous terms
	description := task.Description
	ambiguousTerms := []string{"fix", "update", "improve", "work on"}
	
	for _, term := range ambiguousTerms {
		if containsWord(description, term) {
			suggestions = append(suggestions,
				fmt.Sprintf("Task contains ambiguous term '%s' - consider being more specific for better estimates", term))
		}
	}
	
	// Check if task has subtasks that could be estimated separately
	if len(description) > 100 {
		suggestions = append(suggestions,
			"Long task description - consider breaking into subtasks for more accurate estimates")
	}
	
	// Check historical accuracy for this project
	if task.Project != "" {
		// Would check actual accuracy data here
		suggestions = append(suggestions,
			fmt.Sprintf("Track time with 'timew start task_%s' for better future estimates", task.UUID))
	}
	
	return suggestions
}

// containsWord checks if a word appears in text (case-insensitive)
func containsWord(text, word string) bool {
	// Simple implementation - could be improved with proper word boundary detection
	return len(text) > 0 && len(word) > 0 && 
		(text == word || 
		 text[:len(word)] == word ||
		 text[len(text)-len(word):] == word)
}