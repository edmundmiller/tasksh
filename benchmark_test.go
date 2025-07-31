package main

import (
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
)

func BenchmarkGetTasksOldWay(b *testing.B) {
	// Ensure review config is set up
	if err := taskwarrior.EnsureReviewConfig(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get UUIDs
		uuids, err := taskwarrior.GetTasksForReview()
		if err != nil {
			b.Fatal(err)
		}

		// Limit to 10 tasks for consistent benchmark
		if len(uuids) > 10 {
			uuids = uuids[:10]
		}

		// Get each task individually (old way)
		for _, uuid := range uuids {
			_, err := taskwarrior.GetTaskInfo(uuid)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkGetTasksNewWay(b *testing.B) {
	// Ensure review config is set up
	if err := taskwarrior.EnsureReviewConfig(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get all tasks with data in one call
		tasks, err := taskwarrior.GetTasksForReviewWithData()
		if err != nil {
			b.Fatal(err)
		}

		// Just access the data to simulate usage
		for i, task := range tasks {
			if i >= 10 {
				break
			}
			_ = task.Description
		}
	}
}

func TestPerformanceComparison(t *testing.T) {
	// Ensure review config is set up
	if err := taskwarrior.EnsureReviewConfig(); err != nil {
		t.Fatal(err)
	}

	// Test old way
	start := time.Now()
	uuids, err := taskwarrior.GetTasksForReview()
	if err != nil {
		t.Fatal(err)
	}
	
	// Limit to 10 tasks
	if len(uuids) > 10 {
		uuids = uuids[:10]
	}

	for _, uuid := range uuids {
		_, err := taskwarrior.GetTaskInfo(uuid)
		if err != nil {
			t.Fatal(err)
		}
	}
	oldDuration := time.Since(start)

	// Test new way
	start = time.Now()
	tasks, err := taskwarrior.GetTasksForReviewWithData()
	if err != nil {
		t.Fatal(err)
	}
	
	// Access first 10 tasks
	for i, task := range tasks {
		if i >= 10 {
			break
		}
		_ = task.Description
	}
	newDuration := time.Since(start)

	t.Logf("Old way (10 tasks): %v", oldDuration)
	t.Logf("New way (10 tasks): %v", newDuration)
	t.Logf("Speedup: %.2fx", float64(oldDuration)/float64(newDuration))
}