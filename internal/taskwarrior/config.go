package taskwarrior

import "strings"

// GetProjects returns a list of existing projects
func GetProjects() ([]string, error) {
	output, err := executeTask("rc.verbose:nothing", "_projects")
	if err != nil {
		return []string{}, nil // Return empty list if no projects
	}
	
	if output == "" {
		return []string{}, nil
	}
	
	projects := strings.Split(output, "\n")
	// Filter out empty strings
	var filtered []string
	for _, project := range projects {
		if strings.TrimSpace(project) != "" {
			filtered = append(filtered, strings.TrimSpace(project))
		}
	}
	return filtered, nil
}

// GetTags returns a list of existing tags
func GetTags() ([]string, error) {
	output, err := executeTask("rc.verbose:nothing", "_tags")
	if err != nil {
		return []string{}, nil // Return empty list if no tags
	}
	
	if output == "" {
		return []string{}, nil
	}
	
	tags := strings.Split(output, "\n")
	// Filter out empty strings and add + prefix for suggestions
	var filtered []string
	for _, tag := range tags {
		if strings.TrimSpace(tag) != "" {
			filtered = append(filtered, "+"+strings.TrimSpace(tag))
		}
	}
	return filtered, nil
}

// GetPriorities returns available priority levels
func GetPriorities() []string {
	return []string{"priority:H", "priority:M", "priority:L", "priority:"}
}

// GetCommonModifications returns common modification patterns
func GetCommonModifications() []string {
	return []string{
		"due:tomorrow",
		"due:next week", 
		"due:next month",
		"due:",
		"wait:tomorrow",
		"wait:next week",
		"wait:",
		"scheduled:tomorrow",
		"scheduled:next week",
		"scheduled:",
		"depends:",
	}
}

// GetWaitPeriods returns common wait periods for autocompletion
func GetWaitPeriods() []string {
	return []string{
		"tomorrow",
		"next week",
		"next month",
		"1week",
		"2weeks", 
		"1month",
		"3months",
		"monday",
		"friday",
		"january",
		"next year",
	}
}