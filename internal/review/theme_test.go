package review

import (
	"os"
	"testing"
)

// TestGetTheme tests theme selection
func TestGetTheme(t *testing.T) {
	// Test default theme
	theme := GetTheme()
	if theme == nil {
		t.Fatal("GetTheme returned nil")
	}
	
	if theme.Styles == nil {
		t.Error("Theme styles not initialized")
	}
}

// TestNoColorTheme tests NO_COLOR environment variable
func TestNoColorTheme(t *testing.T) {
	// Save original value
	original := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", original)
	
	// Set NO_COLOR
	os.Setenv("NO_COLOR", "1")
	
	theme := GetTheme()
	if theme.Name != "no-color" {
		t.Errorf("Expected no-color theme, got %s", theme.Name)
	}
}

// TestExplicitTheme tests TASKSH_THEME environment variable
func TestExplicitTheme(t *testing.T) {
	// Save original values
	originalTheme := os.Getenv("TASKSH_THEME")
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TASKSH_THEME", originalTheme)
		os.Setenv("NO_COLOR", originalNoColor)
	}()
	
	// Clear NO_COLOR
	os.Unsetenv("NO_COLOR")
	
	// Test each theme
	themes := []string{"light", "dark", "auto"}
	for _, themeName := range themes {
		os.Setenv("TASKSH_THEME", themeName)
		theme := GetTheme()
		
		if theme == nil {
			t.Errorf("GetTheme returned nil for theme %s", themeName)
			continue
		}
		
		if theme.Styles == nil {
			t.Errorf("Theme %s has nil styles", themeName)
		}
	}
}

// TestThemeStyles tests that each theme has proper styles
func TestThemeStyles(t *testing.T) {
	themes := []*Theme{
		DefaultTheme(),
		DarkTheme(),
		LightTheme(),
		NoColorTheme(),
	}
	
	for _, theme := range themes {
		if theme == nil {
			t.Error("Theme is nil")
			continue
		}
		
		if theme.Name == "" {
			t.Error("Theme has no name")
		}
		
		if theme.Styles == nil {
			t.Errorf("Theme %s has nil styles", theme.Name)
			continue
		}
		
		// Test that key styles are defined
		// We can't easily test the actual style values, but we can
		// ensure they're not zero values
		styles := theme.Styles
		
		// Check that styles have been configured
		// (This is a basic check - in real tests you might want to
		// check specific properties)
		if styles.Title.String() == "" {
			t.Errorf("Theme %s has empty Title style", theme.Name)
		}
		
		if styles.Error.String() == "" {
			t.Errorf("Theme %s has empty Error style", theme.Name)
		}
	}
}

// TestGetThemeInfo tests theme info generation
func TestGetThemeInfo(t *testing.T) {
	// Save original values
	originalTheme := os.Getenv("TASKSH_THEME")
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TASKSH_THEME", originalTheme)
		os.Setenv("NO_COLOR", originalNoColor)
	}()
	
	// Test auto-detected
	os.Unsetenv("TASKSH_THEME")
	os.Unsetenv("NO_COLOR")
	info := GetThemeInfo()
	if !contains(info, "auto-detected") {
		t.Error("Expected auto-detected in theme info")
	}
	
	// Test explicit theme
	os.Setenv("TASKSH_THEME", "dark")
	info = GetThemeInfo()
	if !contains(info, "TASKSH_THEME=dark") {
		t.Error("Expected TASKSH_THEME=dark in theme info")
	}
	
	// Test NO_COLOR
	os.Setenv("NO_COLOR", "1")
	info = GetThemeInfo()
	if !contains(info, "NO_COLOR set") {
		t.Error("Expected NO_COLOR set in theme info")
	}
}