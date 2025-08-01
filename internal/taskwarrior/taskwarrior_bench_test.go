package taskwarrior

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// BenchmarkParseTask benchmarks task parsing from JSON
func BenchmarkParseTask(b *testing.B) {
	jsonData := []string{
		`{"uuid":"123","description":"Simple task"}`,
		`{"uuid":"456","description":"Complex task","project":"work","tags":["urgent","backend"],"priority":"H","due":"20241225T000000Z"}`,
		`{"uuid":"789","description":"Task with annotations","annotations":[{"entry":"20240101T120000Z","description":"First note"},{"entry":"20240102T120000Z","description":"Second note"}]}`,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := jsonData[i%len(jsonData)]
		parseTask([]byte(data))
	}
}

// BenchmarkFormatDue benchmarks due date formatting
func BenchmarkFormatDue(b *testing.B) {
	testCases := []struct {
		name string
		due  string
	}{
		{"Empty", ""},
		{"Today", time.Now().Format("20060102T150405Z")},
		{"Tomorrow", time.Now().AddDate(0, 0, 1).Format("20060102T150405Z")},
		{"NextWeek", time.Now().AddDate(0, 0, 7).Format("20060102T150405Z")},
		{"PastDue", time.Now().AddDate(0, 0, -3).Format("20060102T150405Z")},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			task := &Task{Due: tc.due}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = task.FormatDue()
			}
		})
	}
}

// BenchmarkGetDisplayTags benchmarks tag display formatting
func BenchmarkGetDisplayTags(b *testing.B) {
	testCases := []struct {
		name string
		tags []string
	}{
		{"NoTags", []string{}},
		{"SingleTag", []string{"work"}},
		{"MultipleTags", []string{"work", "urgent", "backend", "review", "testing"}},
		{"ManyTags", []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10"}},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			task := &Task{Tags: tc.tags}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = task.GetDisplayTags()
			}
		})
	}
}

// BenchmarkGetShortDescription benchmarks description truncation
func BenchmarkGetShortDescription(b *testing.B) {
	testCases := []struct {
		name   string
		desc   string
		maxLen int
	}{
		{"Short", "Short task", 20},
		{"ExactLength", "This is exactly 20ch", 20},
		{"NeedsTruncation", "This is a very long description that needs to be truncated", 20},
		{"VeryLong", strings.Repeat("Very long description ", 20), 50},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			task := &Task{Description: tc.desc}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = task.GetShortDescription(tc.maxLen)
			}
		})
	}
}

// BenchmarkBatchLoadTasks simulates batch loading performance
// Note: This benchmark creates mock data since it can't call actual taskwarrior
func BenchmarkBatchLoadTasks(b *testing.B) {
	// Create mock UUID lists of different sizes
	uuidCounts := []int{5, 10, 20, 50}
	
	for _, count := range uuidCounts {
		b.Run(fmt.Sprintf("%dTasks", count), func(b *testing.B) {
			uuids := make([]string, count)
			for i := 0; i < count; i++ {
				uuids[i] = fmt.Sprintf("uuid-%d", i)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate the work of parsing tasks
				// In real usage, this would call taskwarrior
				result := make(map[string]*Task, len(uuids))
				for _, uuid := range uuids {
					result[uuid] = &Task{
						UUID:        uuid,
						Description: "Mock task for " + uuid,
						Project:     "benchmark",
						Tags:        []string{"bench"},
					}
				}
			}
		})
	}
}

// BenchmarkIsOverdue benchmarks overdue checking
func BenchmarkIsOverdue(b *testing.B) {
	now := time.Now()
	testCases := []struct {
		name string
		due  string
	}{
		{"NoDue", ""},
		{"FutureDue", now.AddDate(0, 0, 7).Format("20060102T150405Z")},
		{"TodayDue", now.Format("20060102T150405Z")},
		{"PastDue", now.AddDate(0, 0, -3).Format("20060102T150405Z")},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			task := &Task{Due: tc.due}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = task.IsOverdue()
			}
		})
	}
}

// BenchmarkHasTag benchmarks tag checking
func BenchmarkHasTag(b *testing.B) {
	task := &Task{
		Tags: []string{"work", "urgent", "backend", "review", "testing", "documentation", "refactor"},
	}
	
	testCases := []struct {
		name   string
		tag    string
		exists bool
	}{
		{"ExistsFirst", "work", true},
		{"ExistsLast", "refactor", true},
		{"ExistsMiddle", "review", true},
		{"NotExists", "frontend", false},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = task.HasTag(tc.tag)
			}
		})
	}
}