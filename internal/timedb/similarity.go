package timedb

import (
	"math"
	"sort"
	"strings"
	"time"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// SimilarityScore represents how similar a task is to another
type SimilarityScore struct {
	Entry      TimeEntry
	Score      float64
	MatchType  string // "exact", "project", "keywords", "tags"
}

// GetSimilarTasksEnhanced finds similar tasks using multiple matching strategies
func (tdb *TimeDB) GetSimilarTasksEnhanced(task *taskwarrior.Task, limit int) ([]SimilarityScore, error) {
	// Get all potential matches with basic filtering
	candidates, err := tdb.getCandidateTasks(task, limit*3)
	if err != nil {
		return nil, err
	}
	
	// Score each candidate
	scores := make([]SimilarityScore, 0, len(candidates))
	for _, candidate := range candidates {
		score := tdb.calculateSimilarity(task, candidate)
		if score.Score > 0 {
			scores = append(scores, score)
		}
	}
	
	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	
	// Return top matches up to limit
	if len(scores) > limit {
		scores = scores[:limit]
	}
	
	return scores, nil
}

// getCandidateTasks retrieves potential matching tasks from the database
func (tdb *TimeDB) getCandidateTasks(task *taskwarrior.Task, limit int) ([]TimeEntry, error) {
	// Build a query that gets tasks with any potential match
	query := `
	SELECT uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at
	FROM time_entries 
	WHERE actual_hours > 0
	  AND (project = ? OR project = '' OR ? = '')
	ORDER BY completed_at DESC
	LIMIT ?
	`
	
	rows, err := tdb.db.Query(query, task.Project, task.Project, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entries []TimeEntry
	for rows.Next() {
		var entry TimeEntry
		err := rows.Scan(
			&entry.UUID,
			&entry.Description,
			&entry.Project,
			&entry.Tags,
			&entry.Priority,
			&entry.EstimatedHours,
			&entry.ActualHours,
			&entry.CompletedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	
	return entries, rows.Err()
}

// calculateSimilarity calculates a similarity score between a task and a time entry
func (tdb *TimeDB) calculateSimilarity(task *taskwarrior.Task, entry TimeEntry) SimilarityScore {
	score := 0.0
	matchType := "none"
	
	// Exact description match (rare but highest confidence)
	if strings.EqualFold(task.Description, entry.Description) {
		score = 1.0
		matchType = "exact"
		return SimilarityScore{Entry: entry, Score: score, MatchType: matchType}
	}
	
	// Project match (high confidence)
	if task.Project != "" && task.Project == entry.Project {
		score += 0.5
		matchType = "project"
	}
	
	// Priority match (small boost)
	if task.Priority != "" && task.Priority == entry.Priority {
		score += 0.1
	}
	
	// Keyword matching
	keywordScore := tdb.calculateKeywordSimilarity(task.Description, entry.Description)
	if keywordScore > 0 {
		score += keywordScore * 0.4 // Weight keyword matching at 40%
		if matchType == "none" {
			matchType = "keywords"
		}
	}
	
	// Tag matching (if we had tags)
	// tagScore := tdb.calculateTagSimilarity(task.Tags, entry.Tags)
	// score += tagScore * 0.2
	
	// Time decay - recent completions are more relevant
	daysSince := time.Since(entry.CompletedAt).Hours() / 24
	if daysSince < 30 {
		score *= 1.0 // Full score for last 30 days
	} else if daysSince < 90 {
		score *= 0.8 // 80% for 30-90 days
	} else if daysSince < 180 {
		score *= 0.6 // 60% for 90-180 days
	} else {
		score *= 0.4 // 40% for older
	}
	
	return SimilarityScore{Entry: entry, Score: score, MatchType: matchType}
}

// calculateKeywordSimilarity calculates similarity based on shared keywords
func (tdb *TimeDB) calculateKeywordSimilarity(desc1, desc2 string) float64 {
	// Tokenize and normalize
	words1 := tokenizeDescription(desc1)
	words2 := tokenizeDescription(desc2)
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}
	
	// Create word sets
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}
	
	// Count matches
	matches := 0
	for _, word := range words2 {
		if wordSet1[word] {
			matches++
		}
	}
	
	// Jaccard similarity coefficient
	union := len(wordSet1)
	for _, word := range words2 {
		if !wordSet1[word] {
			union++
		}
	}
	
	if union == 0 {
		return 0
	}
	
	return float64(matches) / float64(union)
}

// tokenizeDescription splits a description into normalized words
func tokenizeDescription(desc string) []string {
	// Convert to lowercase and split
	desc = strings.ToLower(desc)
	words := strings.Fields(desc)
	
	// Filter out common words and clean up
	filtered := make([]string, 0, len(words))
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "with": true, "by": true, "from": true, "as": true,
		"is": true, "was": true, "are": true, "were": true, "been": true,
	}
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()-[]{}/*")
		
		// Skip short words and stop words
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}
	
	return filtered
}

// EstimateBasedOnSimilarity provides time estimation using enhanced similarity matching
func (tdb *TimeDB) EstimateBasedOnSimilarity(task *taskwarrior.Task) (float64, string, float64, error) {
	// Get similar tasks with scores
	similar, err := tdb.GetSimilarTasksEnhanced(task, 10)
	if err != nil {
		return 0, "", 0, err
	}
	
	if len(similar) == 0 {
		return 0, "No similar tasks found", 0, nil
	}
	
	// Calculate weighted average based on similarity scores
	totalWeight := 0.0
	weightedSum := 0.0
	matchTypes := make(map[string]int)
	
	for _, match := range similar {
		weight := match.Score
		weightedSum += match.Entry.ActualHours * weight
		totalWeight += weight
		matchTypes[match.MatchType]++
	}
	
	if totalWeight == 0 {
		return 0, "No weighted matches", 0, nil
	}
	
	estimate := weightedSum / totalWeight
	confidence := calculateConfidence(similar, matchTypes)
	
	// Build reason string
	reason := buildEstimateReason(similar, matchTypes)
	
	return estimate, reason, confidence, nil
}

// calculateConfidence calculates confidence score (0-1) based on match quality
func calculateConfidence(matches []SimilarityScore, matchTypes map[string]int) float64 {
	if len(matches) == 0 {
		return 0
	}
	
	confidence := 0.0
	
	// More matches = higher confidence
	matchCountScore := math.Min(float64(len(matches))/5.0, 1.0) * 0.3
	confidence += matchCountScore
	
	// Better match types = higher confidence
	if matchTypes["exact"] > 0 {
		confidence += 0.4
	} else if matchTypes["project"] > 0 {
		confidence += 0.3
	} else if matchTypes["keywords"] > 0 {
		confidence += 0.2
	}
	
	// Higher average similarity score = higher confidence
	avgScore := 0.0
	for _, match := range matches {
		avgScore += match.Score
	}
	avgScore /= float64(len(matches))
	confidence += avgScore * 0.3
	
	return math.Min(confidence, 1.0)
}

// buildEstimateReason creates a human-readable reason for the estimate
func buildEstimateReason(matches []SimilarityScore, matchTypes map[string]int) string {
	parts := []string{}
	
	if matchTypes["exact"] > 0 {
		parts = append(parts, "exact matches")
	}
	if matchTypes["project"] > 0 {
		parts = append(parts, "same project")
	}
	if matchTypes["keywords"] > 0 {
		parts = append(parts, "similar keywords")
	}
	
	matchDesc := strings.Join(parts, " and ")
	if matchDesc == "" {
		matchDesc = "similar characteristics"
	}
	
	return strings.Title(matchDesc) + " from " + 
		strings.Title(formatInt(len(matches))) + " similar tasks"
}

// formatInt converts numbers to words for better readability
func formatInt(n int) string {
	switch n {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	default:
		return strings.Title(string(rune(n)))
	}
}