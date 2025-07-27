package ai

import (
	"fmt"
	"os/exec"
)

// CheckModsAvailable checks if the mods command is available
func CheckModsAvailable() error {
	cmd := exec.Command("mods", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mods command not found: %w", err)
	}
	return nil
}