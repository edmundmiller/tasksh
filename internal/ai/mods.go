package ai

import (
	"fmt"
	"os"
)

// CheckOpenAIAvailable checks if OpenAI API key is available
func CheckOpenAIAvailable() error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Try the 1Password CLI command from the user's instructions
		if cmd := os.Getenv("OPENAI_API_KEY_CMD"); cmd != "" {
			// We have a command to get the key, that's good enough
			return nil
		}
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	return nil
}