package cli

import (
	"fmt"
	"time"

	"github.com/emiller/tasksh/internal/ai"
	"github.com/emiller/tasksh/internal/estimation"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
	"github.com/emiller/tasksh/internal/timewarrior"
)

func ShowDiagnostics() {
	fmt.Println("tasksh diagnostics")
	fmt.Println()
	fmt.Printf("Version: %s\n", "2.0.0-go")
	fmt.Printf("Built with: Go\n")
	fmt.Println()
	
	// Check if task command is available
	if err := taskwarrior.CheckAvailable(); err != nil {
		fmt.Printf("Taskwarrior: NOT FOUND - %v\n", err)
	} else {
		fmt.Println("Taskwarrior: Available")
	}
	
	// Check if OpenAI API is available  
	if err := ai.CheckOpenAIAvailable(); err != nil {
		fmt.Printf("Mods (AI): NOT AVAILABLE - %v\n", err)
		fmt.Println("  Set OPENAI_API_KEY environment variable or use: export OPENAI_API_KEY=$(op read \"op://Private/api.openai.com/apikey\")")
	} else {
		fmt.Println("Mods (AI): Available")
	}
	
	// Check time database
	if db, err := timedb.New(); err != nil {
		fmt.Printf("Time Database: ERROR - %v\n", err)
	} else {
		defer db.Close()
		fmt.Println("Time Database: Available")
		
		// Check timewarrior
		tw := timewarrior.NewClient()
		if entries, err := tw.Export("2024-01-01", "-", "2024-01-02"); err != nil {
			fmt.Printf("Timewarrior: NOT AVAILABLE - %v\n", err)
		} else {
			fmt.Printf("Timewarrior: Available (test returned %d entries)\n", len(entries))
		}
		
		// Check sync status
		if lastSync, err := db.GetLastSyncTime(); err == nil && !lastSync.IsZero() {
			fmt.Printf("Last Sync: %s (%s ago)\n", 
				lastSync.Format("2006-01-02 15:04:05"),
				time.Since(lastSync).Round(time.Minute))
		} else {
			fmt.Println("Last Sync: Never")
		}
		
		// Show auto-sync configuration
		config := estimation.DefaultConfig()
		fmt.Printf("\nAuto-sync Configuration:\n")
		fmt.Printf("  Enabled: %v\n", config.AutoSyncEnabled)
		fmt.Printf("  Interval: %v\n", config.AutoSyncInterval)
		fmt.Printf("  Next sync: ")
		if lastSync, err := db.GetLastSyncTime(); err == nil && !lastSync.IsZero() {
			nextSync := lastSync.Add(config.AutoSyncInterval)
			if time.Now().After(nextSync) {
				fmt.Println("Due now (will sync on next estimation)")
			} else {
				fmt.Printf("In %v\n", time.Until(nextSync).Round(time.Minute))
			}
		} else {
			fmt.Println("On first estimation")
		}
	}
}