package review

import (
	"sort"
	"strings"

	"github.com/emiller/tasksh/internal/taskwarrior"
)

// CompletionType represents the type of completion being offered
type CompletionType int

const (
	CompletionAttribute CompletionType = iota
	CompletionProject
	CompletionTag
	CompletionPriority
	CompletionStatus
	CompletionValue
)

// CompletionItem represents a single completion suggestion
type CompletionItem struct {
	Text        string
	Description string
	Type        CompletionType
}

// CompletionModel handles auto-completion for modify commands
type CompletionModel struct {
	// Available completion data
	projects   []string
	tags       []string
	attributes []string
	priorities []string
	statuses   []string
	
	// Current completion state
	suggestions []CompletionItem
	input       string
	cursorPos   int
}

// NewCompletionModel creates a new completion model
func NewCompletionModel() *CompletionModel {
	return &CompletionModel{
		attributes: []string{
			"description", "project", "priority", "due", "wait", "scheduled",
			"recur", "until", "depends", "tags", "start", "entry", "end",
			"modified", "status", "urgency", "estimate", "actual_time",
		},
		priorities: []string{"H", "M", "L"},
		statuses:   []string{"pending", "completed", "deleted", "waiting"},
		projects:   []string{}, // Will be loaded dynamically
		tags:       []string{}, // Will be loaded dynamically
	}
}

// LoadDynamicData fetches current projects and tags from taskwarrior
func (c *CompletionModel) LoadDynamicData() error {
	// Get projects
	projects, err := taskwarrior.GetProjects()
	if err == nil {
		c.projects = projects
	}
	
	// Get tags (these come with + prefix, so we need to clean them)
	tagsWithPrefix, err := taskwarrior.GetTags()
	if err == nil {
		c.tags = make([]string, 0, len(tagsWithPrefix))
		for _, tag := range tagsWithPrefix {
			if strings.HasPrefix(tag, "+") {
				c.tags = append(c.tags, tag[1:]) // Remove + prefix
			} else {
				c.tags = append(c.tags, tag)
			}
		}
	}
	
	return nil
}

// UpdateSuggestions updates the completion suggestions based on current input
func (c *CompletionModel) UpdateSuggestions(input string, cursorPos int) {
	c.input = input
	c.cursorPos = cursorPos
	c.suggestions = []CompletionItem{}
	
	if input == "" {
		c.addCommonSuggestions()
		return
	}
	
	// Parse the current input to understand context
	words := strings.Fields(input)
	if len(words) == 0 {
		c.addCommonSuggestions()
		return
	}
	
	lastWord := words[len(words)-1]
	
	// Handle different types of input
	if strings.HasPrefix(lastWord, "+") {
		// Adding a tag
		c.addTagSuggestions(lastWord[1:])
	} else if strings.HasPrefix(lastWord, "-") {
		// Removing a tag
		c.addTagSuggestions(lastWord[1:])
	} else if strings.Contains(lastWord, ":") {
		// Attribute:value format
		parts := strings.SplitN(lastWord, ":", 2)
		if len(parts) == 2 {
			c.addValueSuggestions(parts[0], parts[1])
		}
	} else {
		// Could be start of attribute, tag, or other
		c.addAttributeSuggestions(lastWord)
		c.addTagPrefixSuggestions(lastWord)
		c.addOtherSuggestions(lastWord)
	}
	
	// Sort suggestions by relevance
	c.sortSuggestions()
}

// GetSuggestions returns the current suggestions
func (c *CompletionModel) GetSuggestions() []CompletionItem {
	return c.suggestions
}

// addCommonSuggestions adds the most common modification patterns
func (c *CompletionModel) addCommonSuggestions() {
	common := []CompletionItem{
		{"project:", "Set project", CompletionAttribute},
		{"priority:", "Set priority (H/M/L)", CompletionAttribute},
		{"due:", "Set due date", CompletionAttribute},
		{"+", "Add tag", CompletionTag},
		{"-", "Remove tag", CompletionTag},
		{"description:", "Modify description", CompletionAttribute},
		{"wait:", "Set wait date", CompletionAttribute},
		{"scheduled:", "Set scheduled date", CompletionAttribute},
	}
	c.suggestions = append(c.suggestions, common...)
}

// addAttributeSuggestions adds attribute name suggestions
func (c *CompletionModel) addAttributeSuggestions(prefix string) {
	for _, attr := range c.attributes {
		if strings.HasPrefix(attr, prefix) {
			c.suggestions = append(c.suggestions, CompletionItem{
				Text:        attr + ":",
				Description: "Set " + attr,
				Type:        CompletionAttribute,
			})
		}
	}
}

// addTagSuggestions adds tag suggestions for +tag or -tag format
func (c *CompletionModel) addTagSuggestions(prefix string) {
	for _, tag := range c.tags {
		if strings.HasPrefix(tag, prefix) {
			c.suggestions = append(c.suggestions, CompletionItem{
				Text:        tag,
				Description: "Tag: " + tag,
				Type:        CompletionTag,
			})
		}
	}
}

// addTagPrefixSuggestions adds +tag suggestions when user hasn't typed + yet
func (c *CompletionModel) addTagPrefixSuggestions(prefix string) {
	for _, tag := range c.tags {
		if strings.HasPrefix(tag, prefix) {
			c.suggestions = append(c.suggestions, CompletionItem{
				Text:        "+" + tag,
				Description: "Add tag: " + tag,
				Type:        CompletionTag,
			})
		}
	}
}

// addValueSuggestions adds value suggestions for attribute:value format
func (c *CompletionModel) addValueSuggestions(attribute, valuePrefix string) {
	switch attribute {
	case "project":
		for _, project := range c.projects {
			if strings.HasPrefix(project, valuePrefix) {
				c.suggestions = append(c.suggestions, CompletionItem{
					Text:        project,
					Description: "Project: " + project,
					Type:        CompletionProject,
				})
			}
		}
	case "priority":
		for _, priority := range c.priorities {
			if strings.HasPrefix(priority, valuePrefix) {
				desc := map[string]string{
					"H": "High priority",
					"M": "Medium priority", 
					"L": "Low priority",
				}
				c.suggestions = append(c.suggestions, CompletionItem{
					Text:        priority,
					Description: desc[priority],
					Type:        CompletionPriority,
				})
			}
		}
	case "status":
		for _, status := range c.statuses {
			if strings.HasPrefix(status, valuePrefix) {
				c.suggestions = append(c.suggestions, CompletionItem{
					Text:        status,
					Description: "Status: " + status,
					Type:        CompletionStatus,
				})
			}
		}
	case "due", "wait", "scheduled", "until":
		// Date suggestions
		dateSuggestions := []CompletionItem{
			{"today", "Today", CompletionValue},
			{"tomorrow", "Tomorrow", CompletionValue},
			{"next week", "Next week", CompletionValue},
			{"next month", "Next month", CompletionValue},
			{"eom", "End of month", CompletionValue},
			{"eoy", "End of year", CompletionValue},
		}
		for _, suggestion := range dateSuggestions {
			if strings.HasPrefix(suggestion.Text, valuePrefix) {
				c.suggestions = append(c.suggestions, suggestion)
			}
		}
	}
}

// addOtherSuggestions adds other common modification patterns
func (c *CompletionModel) addOtherSuggestions(prefix string) {
	others := []string{
		"depends:",
		"recur:",
		"estimate:",
	}
	
	for _, other := range others {
		if strings.HasPrefix(other, prefix) {
			c.suggestions = append(c.suggestions, CompletionItem{
				Text:        other,
				Description: "Set " + strings.TrimSuffix(other, ":"),
				Type:        CompletionAttribute,
			})
		}
	}
}

// sortSuggestions sorts suggestions by relevance and type
func (c *CompletionModel) sortSuggestions() {
	sort.Slice(c.suggestions, func(i, j int) bool {
		// First sort by exact prefix match
		iExact := strings.HasPrefix(c.suggestions[i].Text, c.getLastWord())
		jExact := strings.HasPrefix(c.suggestions[j].Text, c.getLastWord())
		
		if iExact && !jExact {
			return true
		}
		if !iExact && jExact {
			return false
		}
		
		// Then by type priority
		typePriority := map[CompletionType]int{
			CompletionAttribute: 1,
			CompletionProject:   2,
			CompletionTag:       3,
			CompletionPriority:  4,
			CompletionStatus:    5,
			CompletionValue:     6,
		}
		
		iPriority := typePriority[c.suggestions[i].Type]
		jPriority := typePriority[c.suggestions[j].Type]
		
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		
		// Finally by alphabetical order
		return c.suggestions[i].Text < c.suggestions[j].Text
	})
	
	// Limit to reasonable number of suggestions
	if len(c.suggestions) > 10 {
		c.suggestions = c.suggestions[:10]
	}
}

// getLastWord gets the word being currently typed
func (c *CompletionModel) getLastWord() string {
	words := strings.Fields(c.input)
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

// GetCompletionText returns the text that should replace the current word
func (c *CompletionModel) GetCompletionText(suggestion CompletionItem, currentInput string) string {
	words := strings.Fields(currentInput)
	if len(words) == 0 {
		return suggestion.Text
	}
	
	lastWord := words[len(words)-1]
	
	// Handle attribute:value completions
	if strings.Contains(lastWord, ":") {
		parts := strings.SplitN(lastWord, ":", 2)
		if len(parts) == 2 {
			attribute := parts[0]
			// For value completions, combine attribute with suggestion
			if suggestion.Type == CompletionProject || 
			   suggestion.Type == CompletionPriority || 
			   suggestion.Type == CompletionStatus || 
			   suggestion.Type == CompletionValue {
				words[len(words)-1] = attribute + ":" + suggestion.Text
				result := strings.Join(words, " ")
				// Add space after completion for easier chaining
				return result + " "
			}
		}
	}
	
	// Handle tag completions
	if strings.HasPrefix(lastWord, "+") || strings.HasPrefix(lastWord, "-") {
		if suggestion.Type == CompletionTag {
			prefix := string(lastWord[0]) // Keep + or -
			words[len(words)-1] = prefix + suggestion.Text
			result := strings.Join(words, " ")
			// Add space after completion for easier chaining
			return result + " "
		}
	}
	
	// Default: replace the last word with the suggestion
	words[len(words)-1] = suggestion.Text
	result := strings.Join(words, " ")
	// Add space after most completions for easier chaining
	if strings.HasSuffix(suggestion.Text, ":") {
		// Don't add space after attribute names that end with :
		return result
	}
	return result + " "
}