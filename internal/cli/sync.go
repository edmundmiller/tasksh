package cli

import (
	"fmt"
	"time"
	
	"github.com/emiller/tasksh/internal/timedb"
	"github.com/emiller/tasksh/internal/timewarrior"
)

// SyncCommand handles the sync command
type SyncCommand struct{}

// Name returns the command name
func (c *SyncCommand) Name() string {
	return "sync"
}

// Description returns the command description
func (c *SyncCommand) Description() string {
	return "Sync time tracking data from timewarrior"
}

// Run executes the sync command
func (c *SyncCommand) Run(args []string) error {
	// Open time database
	db, err := timedb.New()
	if err != nil {
		return fmt.Errorf("failed to open time database: %w", err)
	}
	defer db.Close()
	
	// Create timewarrior client
	tw := timewarrior.NewClient()
	
	// Default to syncing last 30 days
	since := time.Now().AddDate(0, 0, -30)
	
	// Parse args for custom date range
	if len(args) > 0 {
		switch args[0] {
		case "all":
			since = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		case "week":
			since = time.Now().AddDate(0, 0, -7)
		case "month":
			since = time.Now().AddDate(0, -1, 0)
		case "year":
			since = time.Now().AddDate(-1, 0, 0)
		default:
			// Try to parse as date
			parsed, err := time.Parse("2006-01-02", args[0])
			if err != nil {
				return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", args[0])
			}
			since = parsed
		}
	}
	
	fmt.Printf("Syncing timewarrior data since %s...\n", since.Format("2006-01-02"))
	
	// Perform sync
	result, err := db.SyncFromTimewarrior(tw, since)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	
	// Display results
	fmt.Printf("\nSync completed successfully!\n")
	fmt.Printf("- Processed %d entries\n", result.TotalProcessed)
	fmt.Printf("- New entries: %d\n", result.NewEntries)
	fmt.Printf("- Updated entries: %d\n", result.UpdatedEntries)
	
	// Show stats
	stats, err := db.GetTimewarriorStats("")
	if err == nil {
		fmt.Printf("\nDatabase statistics:\n")
		fmt.Printf("- Total unique tasks: %d\n", stats["unique_tasks"])
		fmt.Printf("- Total time entries: %d\n", stats["total_entries"])
		fmt.Printf("- Total hours tracked: %.1f\n", stats["total_hours"])
		
		// Show projects with time
		projects, err := db.GetProjectsWithTime()
		if err == nil && len(projects) > 0 {
			fmt.Printf("\nProjects with tracked time:\n")
			for _, project := range projects {
				fmt.Printf("  - %s\n", project)
			}
		}
	}
	
	return nil
}