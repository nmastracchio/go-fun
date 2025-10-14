package task

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "Test Description"
	priority := High
	dueDate := time.Now().Add(24 * time.Hour)

	task := NewTask(title, description, priority, dueDate)

	if task.Title != title {
		t.Errorf("Expected title %s, got %s", title, task.Title)
	}

	if task.Description != description {
		t.Errorf("Expected description %s, got %s", description, task.Description)
	}

	if task.Priority != priority {
		t.Errorf("Expected priority %v, got %v", priority, task.Priority)
	}

	if !task.DueDate.Equal(dueDate) {
		t.Errorf("Expected due date %v, got %v", dueDate, task.DueDate)
	}

	if task.Completed {
		t.Error("Expected task to be incomplete")
	}

	if task.ID == "" {
		t.Error("Expected task to have an ID")
	}

	if task.CreatedAt.IsZero() {
		t.Error("Expected task to have a creation time")
	}

	if task.UpdatedAt.IsZero() {
		t.Error("Expected task to have an update time")
	}
}

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: &Task{
				Title:       "Valid Task",
				Description: "Valid Description",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			task: &Task{
				Title:       "",
				Description: "Valid Description",
			},
			wantErr: true,
		},
		{
			name: "title too long",
			task: &Task{
				Title:       string(make([]byte, 101)), // 101 characters
				Description: "Valid Description",
			},
			wantErr: true,
		},
		{
			name: "description too long",
			task: &Task{
				Title:       "Valid Title",
				Description: string(make([]byte, 501)), // 501 characters
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskComplete(t *testing.T) {
	task := &Task{
		ID:        "test-id",
		Title:     "Test Task",
		Completed: false,
	}

	originalUpdatedAt := task.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	task.Complete()

	if !task.Completed {
		t.Error("Expected task to be completed")
	}

	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskUncomplete(t *testing.T) {
	task := &Task{
		ID:        "test-id",
		Title:     "Test Task",
		Completed: true,
	}

	originalUpdatedAt := task.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	task.Uncomplete()

	if task.Completed {
		t.Error("Expected task to be uncompleted")
	}

	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskUpdate(t *testing.T) {
	task := &Task{
		ID:          "test-id",
		Title:       "Old Title",
		Description: "Old Description",
		Priority:    Low,
		DueDate:     time.Now(),
		CreatedAt:   time.Now(),
	}

	newTitle := "New Title"
	newDescription := "New Description"
	newPriority := High
	newDueDate := time.Now().Add(48 * time.Hour)

	originalCreatedAt := task.CreatedAt
	originalUpdatedAt := task.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)

	err := task.Update(newTitle, newDescription, newPriority, newDueDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, task.Title)
	}

	if task.Description != newDescription {
		t.Errorf("Expected description %s, got %s", newDescription, task.Description)
	}

	if task.Priority != newPriority {
		t.Errorf("Expected priority %v, got %v", newPriority, task.Priority)
	}

	if !task.DueDate.Equal(newDueDate) {
		t.Errorf("Expected due date %v, got %v", newDueDate, task.DueDate)
	}

	if !task.CreatedAt.Equal(originalCreatedAt) {
		t.Error("Expected CreatedAt to remain unchanged")
	}

	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskIsOverdue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		task     *Task
		expected bool
	}{
		{
			name: "completed task",
			task: &Task{
				Completed: true,
				DueDate:   now.Add(-24 * time.Hour), // Overdue
			},
			expected: false,
		},
		{
			name: "incomplete overdue task",
			task: &Task{
				Completed: false,
				DueDate:   now.Add(-24 * time.Hour), // Overdue
			},
			expected: true,
		},
		{
			name: "incomplete future task",
			task: &Task{
				Completed: false,
				DueDate:   now.Add(24 * time.Hour), // Future
			},
			expected: false,
		},
		{
			name: "no due date",
			task: &Task{
				Completed: false,
				DueDate:   time.Time{}, // Zero time
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.IsOverdue()
			if result != tt.expected {
				t.Errorf("IsOverdue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTaskIsDueToday(t *testing.T) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	tests := []struct {
		name     string
		task     *Task
		expected bool
	}{
		{
			name: "due today",
			task: &Task{
				DueDate: today,
			},
			expected: true,
		},
		{
			name: "due tomorrow",
			task: &Task{
				DueDate: today.Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "due yesterday",
			task: &Task{
				DueDate: today.Add(-24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "no due date",
			task: &Task{
				DueDate: time.Time{}, // Zero time
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.IsDueToday()
			if result != tt.expected {
				t.Errorf("IsDueToday() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTaskIsDueSoon(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		task     *Task
		expected bool
	}{
		{
			name: "due in 3 days",
			task: &Task{
				Completed: false,
				DueDate:   now.Add(3 * 24 * time.Hour),
			},
			expected: true,
		},
		{
			name: "due in 10 days",
			task: &Task{
				Completed: false,
				DueDate:   now.Add(10 * 24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "completed task",
			task: &Task{
				Completed: true,
				DueDate:   now.Add(3 * 24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "overdue task",
			task: &Task{
				Completed: false,
				DueDate:   now.Add(-3 * 24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "no due date",
			task: &Task{
				Completed: false,
				DueDate:   time.Time{}, // Zero time
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.IsDueSoon()
			if result != tt.expected {
				t.Errorf("IsDueSoon() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{Low, "Low"},
		{Medium, "Medium"},
		{High, "High"},
		{Priority(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.priority.String()
			if result != tt.expected {
				t.Errorf("Priority.String() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewTask(b *testing.B) {
	title := "Benchmark Task"
	description := "Benchmark Description"
	priority := Medium
	dueDate := time.Now().Add(24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTask(title, description, priority, dueDate)
	}
}

func BenchmarkTaskValidate(b *testing.B) {
	task := &Task{
		Title:       "Valid Task",
		Description: "Valid Description",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = task.Validate()
	}
}

func BenchmarkTaskComplete(b *testing.B) {
	task := &Task{
		ID:        "test-id",
		Title:     "Test Task",
		Completed: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task.Complete()
		task.Uncomplete() // Reset for next iteration
	}
}
