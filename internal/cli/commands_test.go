package cli

import (
	"context"
	"testing"
	"time"

	"go-fun/internal/storage"
	"go-fun/internal/task"
)

func TestTaskManagerAdd(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	title := "Test Task"
	description := "Test Description"
	priority := task.High
	dueDate := time.Now().Add(24 * time.Hour)

	err := tm.Add(ctx, title, description, priority, dueDate)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Verify task was added
	tasks, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Title != title {
		t.Errorf("Expected title %s, got %s", title, tasks[0].Title)
	}

	if tasks[0].Description != description {
		t.Errorf("Expected description %s, got %s", description, tasks[0].Description)
	}

	if tasks[0].Priority != priority {
		t.Errorf("Expected priority %v, got %v", priority, tasks[0].Priority)
	}
}

func TestTaskManagerComplete(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add a task first
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Complete the task
	err = tm.Complete(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error completing task: %v", err)
	}

	// Verify task is completed
	retrievedTask, err := storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting task: %v", err)
	}

	if !retrievedTask.Completed {
		t.Error("Expected task to be completed")
	}
}

func TestTaskManagerUncomplete(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add a completed task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.Medium,
		Completed:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Uncomplete the task
	err = tm.Uncomplete(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error uncompleting task: %v", err)
	}

	// Verify task is not completed
	retrievedTask, err := storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting task: %v", err)
	}

	if retrievedTask.Completed {
		t.Error("Expected task to be uncompleted")
	}
}

func TestTaskManagerDelete(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add a task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Delete the task
	err = tm.Delete(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error deleting task: %v", err)
	}

	// Verify task is deleted
	tasks, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

func TestTaskManagerUpdate(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add a task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Original Title",
		Description: "Original Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Update the task
	newTitle := "Updated Title"
	newDescription := "Updated Description"
	newPriority := task.High
	newDueDate := time.Now().Add(48 * time.Hour)

	err = tm.Update(ctx, testTask.ID, newTitle, newDescription, newPriority, newDueDate)
	if err != nil {
		t.Fatalf("Unexpected error updating task: %v", err)
	}

	// Verify task is updated
	retrievedTask, err := storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting updated task: %v", err)
	}

	if retrievedTask.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, retrievedTask.Title)
	}

	if retrievedTask.Description != newDescription {
		t.Errorf("Expected description %s, got %s", newDescription, retrievedTask.Description)
	}

	if retrievedTask.Priority != newPriority {
		t.Errorf("Expected priority %v, got %v", newPriority, retrievedTask.Priority)
	}
}

func TestTaskManagerShow(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add a task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.High,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Show the task (this should not return an error)
	err = tm.Show(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error showing task: %v", err)
	}
}

func TestTaskManagerStats(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Add some test tasks
	tasks := []*task.Task{
		{
			ID:          "test-1",
			Title:       "High Priority Task",
			Description: "High priority",
			Priority:    task.High,
			Completed:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "test-2",
			Title:       "Medium Priority Task",
			Description: "Medium priority",
			Priority:    task.Medium,
			Completed:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "test-3",
			Title:       "Low Priority Task",
			Description: "Low priority",
			Priority:    task.Low,
			Completed:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, task := range tasks {
		err := storage.Add(ctx, task)
		if err != nil {
			t.Fatalf("Unexpected error adding task: %v", err)
		}
	}

	// Get stats (this should not return an error)
	err := tm.Stats(ctx)
	if err != nil {
		t.Fatalf("Unexpected error getting stats: %v", err)
	}
}

func TestTaskManagerErrorHandling(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Test completing non-existent task
	err := tm.Complete(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when completing non-existent task")
	}

	// Test uncompleting non-existent task
	err = tm.Uncomplete(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when uncompleting non-existent task")
	}

	// Test deleting non-existent task
	err = tm.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent task")
	}

	// Test updating non-existent task
	err = tm.Update(ctx, "non-existent", "Title", "Description", task.Medium, time.Now())
	if err == nil {
		t.Error("Expected error when updating non-existent task")
	}

	// Test showing non-existent task
	err = tm.Show(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when showing non-existent task")
	}
}

// Benchmark tests
func BenchmarkTaskManagerAdd(b *testing.B) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	title := "Benchmark Task"
	description := "Benchmark Description"
	priority := task.Medium
	dueDate := time.Now().Add(24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Add(ctx, title, description, priority, dueDate)
	}
}

func BenchmarkTaskManagerComplete(b *testing.B) {
	storage := storage.NewInMemoryStorage()
	tm := NewTaskManager(storage)
	ctx := context.Background()

	// Pre-add a task
	testTask := &task.Task{
		ID:          "benchmark-task",
		Title:       "Benchmark Task",
		Description: "Benchmark Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	storage.Add(ctx, testTask)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Complete(ctx, testTask.ID)
		tm.Uncomplete(ctx, testTask.ID) // Reset for next iteration
	}
}
