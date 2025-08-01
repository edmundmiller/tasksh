package estimation

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	
	_ "modernc.org/sqlite"
)

// AICache provides persistent caching for AI estimates
type AICache struct {
	db *sql.DB
}

// NewAICache creates a new AI cache
func NewAICache(dbPath string) (*AICache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	
	cache := &AICache{db: db}
	if err := cache.initSchema(); err != nil {
		db.Close()
		return nil, err
	}
	
	return cache, nil
}

// Close closes the cache database
func (c *AICache) Close() error {
	return c.db.Close()
}

// initSchema creates the cache table
func (c *AICache) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS ai_estimates (
		task_hash TEXT PRIMARY KEY,
		estimate_hours REAL NOT NULL,
		confidence REAL NOT NULL,
		reason TEXT NOT NULL,
		details TEXT,
		created_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_expires_at ON ai_estimates(expires_at);
	`
	
	_, err := c.db.Exec(schema)
	return err
}

// Get retrieves a cached estimate
func (c *AICache) Get(taskDescription, project string) (*Estimate, error) {
	hash := c.hashTask(taskDescription, project)
	
	query := `
	SELECT estimate_hours, confidence, reason, details
	FROM ai_estimates
	WHERE task_hash = ? AND expires_at > ?
	`
	
	var hours, confidence float64
	var reason string
	var detailsJSON sql.NullString
	
	err := c.db.QueryRow(query, hash, time.Now()).Scan(
		&hours, &confidence, &reason, &detailsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	estimate := &Estimate{
		Hours:      hours,
		Source:     SourceAI,
		Confidence: confidence,
		Reason:     reason + " (cached)",
	}
	
	if detailsJSON.Valid {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(detailsJSON.String), &details); err == nil {
			estimate.Details = details
		}
	}
	
	return estimate, nil
}

// Set stores an estimate in the cache
func (c *AICache) Set(taskDescription, project string, estimate *Estimate, ttl time.Duration) error {
	hash := c.hashTask(taskDescription, project)
	
	detailsJSON := ""
	if estimate.Details != nil {
		data, err := json.Marshal(estimate.Details)
		if err == nil {
			detailsJSON = string(data)
		}
	}
	
	query := `
	INSERT OR REPLACE INTO ai_estimates
	(task_hash, estimate_hours, confidence, reason, details, created_at, expires_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	_, err := c.db.Exec(query,
		hash,
		estimate.Hours,
		estimate.Confidence,
		estimate.Reason,
		detailsJSON,
		now,
		now.Add(ttl),
	)
	
	return err
}

// CleanExpired removes expired cache entries
func (c *AICache) CleanExpired() error {
	query := `DELETE FROM ai_estimates WHERE expires_at < ?`
	_, err := c.db.Exec(query, time.Now())
	return err
}

// hashTask creates a consistent hash for a task
func (c *AICache) hashTask(description, project string) string {
	h := sha256.New()
	h.Write([]byte(description))
	h.Write([]byte("|"))
	h.Write([]byte(project))
	return fmt.Sprintf("%x", h.Sum(nil))
}