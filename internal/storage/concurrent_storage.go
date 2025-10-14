package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go-fun/internal/task"
)

// ConcurrentStorage wraps a Storage with concurrency features
type ConcurrentStorage struct {
	storage Storage
	mutex   sync.RWMutex

	// Background save functionality
	autoSaveEnabled bool
	autoSaveTicker  *time.Ticker
	autoSaveStop    chan struct{}
	unsavedTasks    []*task.Task
	unsavedMutex    sync.Mutex
}

// NewConcurrentStorage creates a new concurrent storage wrapper
func NewConcurrentStorage(s Storage) *ConcurrentStorage {
	return &ConcurrentStorage{
		storage:      s,
		autoSaveStop: make(chan struct{}),
	}
}

// EnableAutoSave enables automatic background saving every interval
func (cs *ConcurrentStorage) EnableAutoSave(interval time.Duration) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.autoSaveEnabled {
		return // Already enabled
	}

	cs.autoSaveEnabled = true
	cs.autoSaveTicker = time.NewTicker(interval)

	go cs.autoSaveWorker()
}

// DisableAutoSave disables automatic background saving
func (cs *ConcurrentStorage) DisableAutoSave() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if !cs.autoSaveEnabled {
		return
	}

	cs.autoSaveEnabled = false
	cs.autoSaveTicker.Stop()
	close(cs.autoSaveStop)
}

// autoSaveWorker runs in the background and saves tasks periodically
func (cs *ConcurrentStorage) autoSaveWorker() {
	for {
		select {
		case <-cs.autoSaveTicker.C:
			cs.saveUnsavedTasks()
		case <-cs.autoSaveStop:
			// Final save before stopping
			cs.saveUnsavedTasks()
			return
		}
	}
}

// saveUnsavedTasks saves any unsaved tasks
func (cs *ConcurrentStorage) saveUnsavedTasks() {
	cs.unsavedMutex.Lock()
	defer cs.unsavedMutex.Unlock()

	if len(cs.unsavedTasks) == 0 {
		return
	}

	// Load current tasks and merge with unsaved ones
	ctx := context.Background()
	currentTasks, err := cs.storage.Load(ctx)
	if err != nil {
		// If we can't load, just save the unsaved tasks
		cs.storage.Save(ctx, cs.unsavedTasks)
		cs.unsavedTasks = nil
		return
	}

	// Merge unsaved tasks with current ones
	mergedTasks := cs.mergeTasks(currentTasks, cs.unsavedTasks)

	// Save merged tasks
	if err := cs.storage.Save(ctx, mergedTasks); err == nil {
		cs.unsavedTasks = nil // Clear unsaved tasks on successful save
	}
}

// mergeTasks merges current tasks with unsaved tasks
func (cs *ConcurrentStorage) mergeTasks(current, unsaved []*task.Task) []*task.Task {
	// Create a map of current tasks by ID for quick lookup
	currentMap := make(map[string]*task.Task)
	for _, t := range current {
		currentMap[t.ID] = t
	}

	// Update or add unsaved tasks
	for _, unsavedTask := range unsaved {
		currentMap[unsavedTask.ID] = unsavedTask
	}

	// Convert back to slice
	result := make([]*task.Task, 0, len(currentMap))
	for _, t := range currentMap {
		result = append(result, t)
	}

	return result
}

// QueueTaskForSave queues a task for background saving
func (cs *ConcurrentStorage) QueueTaskForSave(t *task.Task) {
	if !cs.autoSaveEnabled {
		return
	}

	cs.unsavedMutex.Lock()
	defer cs.unsavedMutex.Unlock()

	// Add or update the task in unsaved list
	found := false
	for i, existing := range cs.unsavedTasks {
		if existing.ID == t.ID {
			cs.unsavedTasks[i] = t
			found = true
			break
		}
	}

	if !found {
		cs.unsavedTasks = append(cs.unsavedTasks, t)
	}
}

// Load implements Storage interface
func (cs *ConcurrentStorage) Load(ctx context.Context) ([]*task.Task, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	tasks, err := cs.storage.Load(ctx)
	if err != nil {
		return nil, err
	}

	// Merge with any unsaved tasks
	cs.unsavedMutex.Lock()
	if len(cs.unsavedTasks) > 0 {
		tasks = cs.mergeTasks(tasks, cs.unsavedTasks)
	}
	cs.unsavedMutex.Unlock()

	return tasks, nil
}

// Save implements Storage interface
func (cs *ConcurrentStorage) Save(ctx context.Context, tasks []*task.Task) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Clear unsaved tasks since we're doing a full save
	cs.unsavedMutex.Lock()
	cs.unsavedTasks = nil
	cs.unsavedMutex.Unlock()

	return cs.storage.Save(ctx, tasks)
}

// Add implements Storage interface
func (cs *ConcurrentStorage) Add(ctx context.Context, t *task.Task) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.autoSaveEnabled {
		cs.QueueTaskForSave(t)
		return nil // Don't save immediately if auto-save is enabled
	}

	return cs.storage.Add(ctx, t)
}

// Update implements Storage interface
func (cs *ConcurrentStorage) Update(ctx context.Context, id string, t *task.Task) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.autoSaveEnabled {
		cs.QueueTaskForSave(t)
		return nil // Don't save immediately if auto-save is enabled
	}

	return cs.storage.Update(ctx, id, t)
}

// Delete implements Storage interface
func (cs *ConcurrentStorage) Delete(ctx context.Context, id string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.autoSaveEnabled {
		// Remove from unsaved tasks if present
		cs.unsavedMutex.Lock()
		for i, t := range cs.unsavedTasks {
			if t.ID == id {
				cs.unsavedTasks = append(cs.unsavedTasks[:i], cs.unsavedTasks[i+1:]...)
				break
			}
		}
		cs.unsavedMutex.Unlock()

		// Mark for deletion by setting a special flag or removing from storage
		// For simplicity, we'll do immediate deletion
		return cs.storage.Delete(ctx, id)
	}

	return cs.storage.Delete(ctx, id)
}

// GetByID implements Storage interface
func (cs *ConcurrentStorage) GetByID(ctx context.Context, id string) (*task.Task, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Check unsaved tasks first
	cs.unsavedMutex.Lock()
	for _, t := range cs.unsavedTasks {
		if t.ID == id {
			cs.unsavedMutex.Unlock()
			return t, nil
		}
	}
	cs.unsavedMutex.Unlock()

	// Check in storage
	return cs.storage.GetByID(ctx, id)
}

// ExportManager handles concurrent exports
type ExportManager struct {
	storage Storage
}

// NewExportManager creates a new export manager
func NewExportManager(s Storage) *ExportManager {
	return &ExportManager{
		storage: s,
	}
}

// ConcurrentExport exports tasks to multiple formats concurrently
func (em *ExportManager) ConcurrentExport(ctx context.Context, formats []string, baseFilename string) error {
	if len(formats) == 0 {
		return fmt.Errorf("no formats specified")
	}

	// Load tasks once
	tasks, err := em.storage.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Create channels for results
	type exportResult struct {
		format string
		err    error
	}

	results := make(chan exportResult, len(formats))

	// Start export goroutines
	for _, format := range formats {
		go func(fmt string) {
			filename := baseFilename + "." + fmt
			err := em.exportFormat(tasks, fmt, filename)
			results <- exportResult{format: fmt, err: err}
		}(format)
	}

	// Collect results
	var errors []string
	for i := 0; i < len(formats); i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.format, result.err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("export errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// exportFormat exports tasks to a specific format
func (em *ExportManager) exportFormat(tasks []*task.Task, format, filename string) error {
	switch strings.ToLower(format) {
	case "json":
		return em.exportJSON(tasks, filename)
	case "csv":
		return em.exportCSV(tasks, filename)
	case "markdown", "md":
		return em.exportMarkdown(tasks, filename)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// Export methods (similar to CLI commands but for direct use)
func (em *ExportManager) exportJSON(tasks []*task.Task, filename string) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func (em *ExportManager) exportCSV(tasks []*task.Task, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write CSV header
	fmt.Fprintln(file, "ID,Title,Description,Priority,Completed,Due Date,Created,Updated")

	// Write task data
	for _, t := range tasks {
		dueDate := ""
		if !t.DueDate.IsZero() {
			dueDate = t.DueDate.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(file, "%s,%s,%s,%s,%t,%s,%s,%s\n",
			t.ID,
			strings.ReplaceAll(t.Title, ",", ";"),
			strings.ReplaceAll(t.Description, ",", ";"),
			t.Priority.String(),
			t.Completed,
			dueDate,
			t.CreatedAt.Format("2006-01-02 15:04"),
			t.UpdatedAt.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

func (em *ExportManager) exportMarkdown(tasks []*task.Task, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create Markdown file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# Task Export\n\n")
	fmt.Fprintf(file, "Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Group tasks by completion status
	completed := make([]*task.Task, 0)
	pending := make([]*task.Task, 0)

	for _, t := range tasks {
		if t.Completed {
			completed = append(completed, t)
		} else {
			pending = append(pending, t)
		}
	}

	// Write pending tasks
	if len(pending) > 0 {
		fmt.Fprintf(file, "## Pending Tasks (%d)\n\n", len(pending))
		for _, t := range pending {
			em.writeMarkdownTask(file, t)
		}
		fmt.Fprintln(file)
	}

	// Write completed tasks
	if len(completed) > 0 {
		fmt.Fprintf(file, "## Completed Tasks (%d)\n\n", len(completed))
		for _, t := range completed {
			em.writeMarkdownTask(file, t)
		}
	}

	return nil
}

func (em *ExportManager) writeMarkdownTask(file *os.File, t *task.Task) {
	status := "❌"
	if t.Completed {
		status = "✅"
	}

	priorityEmoji := ""
	switch t.Priority {
	case task.High:
		priorityEmoji = "🔴"
	case task.Medium:
		priorityEmoji = "🟡"
	case task.Low:
		priorityEmoji = "🟢"
	}

	fmt.Fprintf(file, "### %s %s %s\n\n", status, priorityEmoji, t.Title)

	if t.Description != "" {
		fmt.Fprintf(file, "**Description:** %s\n\n", t.Description)
	}

	if !t.DueDate.IsZero() {
		fmt.Fprintf(file, "**Due:** %s\n\n", t.DueDate.Format("2006-01-02 15:04"))
	}

	fmt.Fprintf(file, "**ID:** `%s`  \n", t.ID)
	fmt.Fprintf(file, "**Created:** %s  \n", t.CreatedAt.Format("2006-01-02 15:04"))
	if t.UpdatedAt.After(t.CreatedAt) {
		fmt.Fprintf(file, "**Updated:** %s  \n", t.UpdatedAt.Format("2006-01-02 15:04"))
	}
	fmt.Fprintln(file)
}
