package shared

import (
	"testing"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestHighPriorityQuickTasks(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	// Examples of high priority tasks that are still quick
	tests := []struct {
		description   string
		priority      string
		expectedHours float64
		scenario      string
	}{
		{
			description:   "Take heart medication",
			priority:      "H",
			expectedHours: 0.1, // 6 minutes
			scenario:      "Critical health task but very quick",
		},
		{
			description:   "Reply to CEO email",
			priority:      "H", 
			expectedHours: 0.25, // 15 minutes
			scenario:      "Urgent communication but still quick",
		},
		{
			description:   "Check server status after alert",
			priority:      "H",
			expectedHours: 0.25, // 15 minutes
			scenario:      "Critical check but quick to do",
		},
		{
			description:   "Call doctor to confirm appointment",
			priority:      "H",
			expectedHours: 0.5, // 30 minutes
			scenario:      "Important call but standard duration",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			task := &taskwarrior.Task{
				Description: tt.description,
				Priority:    tt.priority,
			}
			
			hours := analyzer.fallbackEstimate(task)
			
			if hours < tt.expectedHours-0.01 || hours > tt.expectedHours+0.01 {
				t.Errorf("%s: expected %.2f hours, got %.2f hours", 
					tt.scenario, tt.expectedHours, hours)
			}
		})
	}
}

func TestLowPriorityLongTasks(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	// Examples of low priority tasks that take a long time
	tests := []struct {
		description   string
		priority      string
		expectedHours float64
		scenario      string
	}{
		{
			description:   "Research new JavaScript frameworks for next project",
			priority:      "L",
			expectedHours: 2.5, // 2.5 hours
			scenario:      "Low priority research but still time-consuming",
		},
		{
			description:   "Plan workshop for team building",
			priority:      "L",
			expectedHours: 2.5, // 2.5 hours
			scenario:      "Not urgent but requires significant planning time",
		},
		{
			description:   "Create comprehensive documentation for internal tool",
			priority:      "L",
			expectedHours: 1.5, // 1.5 hours
			scenario:      "Low priority but substantial work",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			task := &taskwarrior.Task{
				Description: tt.description,
				Priority:    tt.priority,
			}
			
			hours := analyzer.fallbackEstimate(task)
			
			if hours < tt.expectedHours-0.01 || hours > tt.expectedHours+0.01 {
				t.Errorf("%s: expected %.2f hours, got %.2f hours", 
					tt.scenario, tt.expectedHours, hours)
			}
		})
	}
}