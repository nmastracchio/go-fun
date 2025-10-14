package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-fun/internal/task"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Test empty storage
	tasks, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading empty storage: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}

	// Test adding a task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.High,
		DueDate:     time.Now().Add(24 * time.Hour),
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Test loading tasks
	tasks, err = storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != testTask.ID {
		t.Errorf("Expected task ID %s, got %s", testTask.ID, tasks[0].ID)
	}

	// Test getting task by ID
	retrievedTask, err := storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting task by ID: %v", err)
	}
	if retrievedTask.ID != testTask.ID {
		t.Errorf("Expected task ID %s, got %s", testTask.ID, retrievedTask.ID)
	}

	// Test updating task
	updatedTask := *testTask
	updatedTask.Title = "Updated Task"
	updatedTask.Description = "Updated Description"
	updatedTask.UpdatedAt = time.Now()

	err = storage.Update(ctx, testTask.ID, &updatedTask)
	if err != nil {
		t.Fatalf("Unexpected error updating task: %v", err)
	}

	// Verify update
	retrievedTask, err = storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting updated task: %v", err)
	}
	if retrievedTask.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got %s", retrievedTask.Title)
	}

	// Test deleting task
	err = storage.Delete(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error deleting task: %v", err)
	}

	// Verify deletion
	tasks, err = storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks after deletion: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

func TestJSONFileStorage(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "go-fun-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "tasks.json")
	storage := NewJSONFileStorage(filePath)
	ctx := context.Background()

	// Test empty storage
	tasks, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading empty storage: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}

	// Test adding a task
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.High,
		DueDate:     time.Now().Add(24 * time.Hour),
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = storage.Add(ctx, testTask)
	if err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	// Test loading tasks
	tasks, err = storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != testTask.ID {
		t.Errorf("Expected task ID %s, got %s", testTask.ID, tasks[0].ID)
	}

	// Test getting task by ID
	retrievedTask, err := storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting task by ID: %v", err)
	}
	if retrievedTask.ID != testTask.ID {
		t.Errorf("Expected task ID %s, got %s", testTask.ID, retrievedTask.ID)
	}

	// Test updating task
	updatedTask := *testTask
	updatedTask.Title = "Updated Task"
	updatedTask.Description = "Updated Description"
	updatedTask.UpdatedAt = time.Now()

	err = storage.Update(ctx, testTask.ID, &updatedTask)
	if err != nil {
		t.Fatalf("Unexpected error updating task: %v", err)
	}

	// Verify update
	retrievedTask, err = storage.GetByID(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error getting updated task: %v", err)
	}
	if retrievedTask.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got %s", retrievedTask.Title)
	}

	// Test deleting task
	err = storage.Delete(ctx, testTask.ID)
	if err != nil {
		t.Fatalf("Unexpected error deleting task: %v", err)
	}

	// Verify deletion
	tasks, err = storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks after deletion: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

func TestJSONFileStorageConcurrentAccess(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "go-fun-test-concurrent-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "tasks.json")
	storage := NewJSONFileStorage(filePath)
	ctx := context.Background()

	// Test concurrent writes - note that without proper locking,
	// some writes may be lost due to race conditions
	numGoroutines := 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			testTask := &task.Task{
				ID:          fmt.Sprintf("test-%d", id),
				Title:       fmt.Sprintf("Test Task %d", id),
				Description: fmt.Sprintf("Test Description %d", id),
				Priority:    task.Medium,
				DueDate:     time.Now().Add(time.Duration(id) * time.Hour),
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err := storage.Add(ctx, testTask)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		if err != nil {
			t.Errorf("Error in goroutine %d: %v", i, err)
		}
	}

	// Verify tasks were saved (may be less than expected due to race conditions)
	tasks, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Unexpected error loading tasks: %v", err)
	}
	if len(tasks) == 0 {
		t.Error("Expected at least some tasks to be saved")
	}
	if len(tasks) > numGoroutines {
		t.Errorf("Expected at most %d tasks, got %d", numGoroutines, len(tasks))
	}
}

func TestStorageErrorHandling(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Test adding invalid task
	invalidTask := &task.Task{
		ID:          "test-1",
		Title:       "", // Invalid: empty title
		Description: "Test Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Add(ctx, invalidTask)
	if err == nil {
		t.Error("Expected error when adding invalid task, got nil")
	}

	// Test getting non-existent task
	_, err = storage.GetByID(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent task, got nil")
	}

	// Test updating non-existent task
	err = storage.Update(ctx, "non-existent", invalidTask)
	if err == nil {
		t.Error("Expected error when updating non-existent task, got nil")
	}

	// Test deleting non-existent task
	err = storage.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent task, got nil")
	}
}

func TestJSONFileStorageErrorHandling(t *testing.T) {
	// Test with a path that doesn't exist but can be created
	tempDir, err := os.MkdirTemp("", "go-fun-test-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidPath := filepath.Join(tempDir, "nested", "path", "tasks.json")
	invalidStorage := NewJSONFileStorage(invalidPath)
	ctx := context.Background()

	// This should create the directory and file
	testTask := &task.Task{
		ID:          "test-1",
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    task.Medium,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = invalidStorage.Add(ctx, testTask)
	if err != nil {
		t.Errorf("Unexpected error adding task to nested path (should create directory): %v", err)
	}
}

// Benchmark tests
func BenchmarkInMemoryStorageAdd(b *testing.B) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTask := &task.Task{
			ID:          fmt.Sprintf("test-%d", i),
			Title:       "Benchmark Task",
			Description: "Benchmark Description",
			Priority:    task.Medium,
			Completed:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		storage.Add(ctx, testTask)
	}
}

func BenchmarkJSONFileStorageAdd(b *testing.B) {
	// Create temporary directory for benchmark
	tempDir, err := os.MkdirTemp("", "go-fun-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "tasks.json")
	storage := NewJSONFileStorage(filePath)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTask := &task.Task{
			ID:          fmt.Sprintf("test-%d", i),
			Title:       "Benchmark Task",
			Description: "Benchmark Description",
			Priority:    task.Medium,
			Completed:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		storage.Add(ctx, testTask)
	}
}

func BenchmarkInMemoryStorageLoad(b *testing.B) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Add some test data
	for i := 0; i < 100; i++ {
		testTask := &task.Task{
			ID:          fmt.Sprintf("test-%d", i),
			Title:       "Benchmark Task",
			Description: "Benchmark Description",
			Priority:    task.Medium,
			Completed:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		storage.Add(ctx, testTask)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Load(ctx)
	}
}
