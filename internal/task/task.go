package task

import (
	"fmt"
	"time"
)

// Priority represents the priority level of a task
type Priority int

const (
	Low Priority = iota
	Medium
	High
)

// String returns the string representation of Priority
func (p Priority) String() string {
	switch p {
	case Low:
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	default:
		return "Unknown"
	}
}

// Task represents a single todo item
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}

// NewTask creates a new task with the given parameters
func NewTask(title, description string, priority Priority, dueDate time.Time, tags []string) *Task {
	now := time.Now()
	return &Task{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Priority:    priority,
		DueDate:     dueDate,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        tags,
	}
}

// Validate checks if the task has valid data
func (t *Task) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("task title cannot be empty")
	}
	if len(t.Title) > 100 {
		return fmt.Errorf("task title cannot exceed 100 characters")
	}
	if len(t.Description) > 500 {
		return fmt.Errorf("task description cannot exceed 500 characters")
	}
	return nil
}

// Complete marks the task as completed
func (t *Task) Complete() {
	t.Completed = true
	t.UpdatedAt = time.Now()
}

// Uncomplete marks the task as not completed
func (t *Task) Uncomplete() {
	t.Completed = false
	t.UpdatedAt = time.Now()
}

// Update updates the task with new information
func (t *Task) Update(title, description string, priority Priority, dueDate time.Time) error {
	t.Title = title
	t.Description = description
	t.Priority = priority
	t.DueDate = dueDate
	t.UpdatedAt = time.Now()

	return t.Validate()
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	return !t.Completed && !t.DueDate.IsZero() && t.DueDate.Before(time.Now())
}

// IsDueToday checks if the task is due today
func (t *Task) IsDueToday() bool {
	if t.DueDate.IsZero() {
		return false
	}

	today := time.Now().Truncate(24 * time.Hour)
	dueDate := t.DueDate.Truncate(24 * time.Hour)

	return today.Equal(dueDate)
}

// IsDueSoon checks if the task is due within the next 7 days
func (t *Task) IsDueSoon() bool {
	if t.DueDate.IsZero() || t.Completed {
		return false
	}

	sevenDaysFromNow := time.Now().Add(7 * 24 * time.Hour)
	return t.DueDate.Before(sevenDaysFromNow) && t.DueDate.After(time.Now())
}

// generateID generates a simple unique ID for the task
func generateID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
