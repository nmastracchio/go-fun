package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-fun/internal/task"
)

// Storage defines the interface for task persistence
type Storage interface {
	Load(ctx context.Context) ([]*task.Task, error)
	Save(ctx context.Context, tasks []*task.Task) error
	Add(ctx context.Context, t *task.Task) error
	Update(ctx context.Context, id string, t *task.Task) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*task.Task, error)
}

// JSONFileStorage implements Storage using JSON file persistence
type JSONFileStorage struct {
	filePath string
	mutex    sync.RWMutex
}

// NewJSONFileStorage creates a new JSON file storage instance
func NewJSONFileStorage(filePath string) *JSONFileStorage {
	return &JSONFileStorage{
		filePath: filePath,
	}
}

// Load loads tasks from the JSON file
func (s *JSONFileStorage) Load(ctx context.Context) ([]*task.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []*task.Task{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", s.filePath, err)
	}

	if len(data) == 0 {
		return []*task.Task{}, nil
	}

	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tasks, nil
}

// Save saves tasks to the JSON file
func (s *JSONFileStorage) Save(ctx context.Context, tasks []*task.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file first, then rename (atomic operation)
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// Add adds a new task to storage
func (s *JSONFileStorage) Add(ctx context.Context, t *task.Task) error {
	tasks, err := s.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Validate the task
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Check for duplicate ID
	for _, existing := range tasks {
		if existing.ID == t.ID {
			return fmt.Errorf("task with ID %s already exists", t.ID)
		}
	}

	tasks = append(tasks, t)
	return s.Save(ctx, tasks)
}

// Update updates an existing task
func (s *JSONFileStorage) Update(ctx context.Context, id string, t *task.Task) error {
	tasks, err := s.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Validate the task
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	found := false
	for i, existing := range tasks {
		if existing.ID == id {
			// Preserve original creation time
			t.CreatedAt = existing.CreatedAt
			t.ID = id // Ensure ID doesn't change
			tasks[i] = t
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return s.Save(ctx, tasks)
}

// Delete deletes a task by ID
func (s *JSONFileStorage) Delete(ctx context.Context, id string) error {
	tasks, err := s.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	found := false
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return s.Save(ctx, tasks)
}

// GetByID retrieves a task by its ID
func (s *JSONFileStorage) GetByID(ctx context.Context, id string) (*task.Task, error) {
	tasks, err := s.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task with ID %s not found", id)
}

// InMemoryStorage is a simple in-memory storage for testing
type InMemoryStorage struct {
	tasks []*task.Task
	mutex sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks: make([]*task.Task, 0),
	}
}

// Load returns all tasks from memory
func (s *InMemoryStorage) Load(ctx context.Context) ([]*task.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modifications
	result := make([]*task.Task, len(s.tasks))
	copy(result, s.tasks)
	return result, nil
}

// Save replaces all tasks in memory
func (s *InMemoryStorage) Save(ctx context.Context, tasks []*task.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store a copy to prevent external modifications
	s.tasks = make([]*task.Task, len(tasks))
	copy(s.tasks, tasks)
	return nil
}

// Add adds a new task to memory
func (s *InMemoryStorage) Add(ctx context.Context, t *task.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Check for duplicate ID
	for _, existing := range s.tasks {
		if existing.ID == t.ID {
			return fmt.Errorf("task with ID %s already exists", t.ID)
		}
	}

	s.tasks = append(s.tasks, t)
	return nil
}

// Update updates an existing task in memory
func (s *InMemoryStorage) Update(ctx context.Context, id string, t *task.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	found := false
	for i, existing := range s.tasks {
		if existing.ID == id {
			t.CreatedAt = existing.CreatedAt
			t.ID = id
			s.tasks[i] = t
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return nil
}

// Delete deletes a task from memory
func (s *InMemoryStorage) Delete(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	found := false
	for i, task := range s.tasks {
		if task.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task with ID %s not found", id)
	}

	return nil
}

// GetByID retrieves a task by ID from memory
func (s *InMemoryStorage) GetByID(ctx context.Context, id string) (*task.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, task := range s.tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task with ID %s not found", id)
}
