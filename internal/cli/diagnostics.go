package cli

import (
	"fmt"

	"github.com/emiller/tasksh/internal/ai"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
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
	
	// Check if mods command is available
	if err := ai.CheckModsAvailable(); err != nil {
		fmt.Printf("Mods (AI): NOT FOUND - %v\n", err)
		fmt.Println("  Install mods for AI-assisted task analysis: https://github.com/charmbracelet/mods")
	} else {
		fmt.Println("Mods (AI): Available")
	}
	
	// Check time database
	if db, err := timedb.New(); err != nil {
		fmt.Printf("Time Database: ERROR - %v\n", err)
	} else {
		db.Close()
		fmt.Println("Time Database: Available")
	}
}