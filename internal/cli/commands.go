package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"go-fun/internal/filter"
	"go-fun/internal/storage"
	"go-fun/internal/task"
	"os"
	"sort"
	"strings"
	"time"
)

// TaskManager handles CLI operations for tasks
type TaskManager struct {
	storage storage.Storage
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(s storage.Storage) *TaskManager {
	return &TaskManager{
		storage: s,
	}
}

// Add creates a new task
func (tm *TaskManager) Add(ctx context.Context, title, description string, priority task.Priority, dueDate time.Time) error {
	newTask := task.NewTask(title, description, priority, dueDate)
	return tm.storage.Add(ctx, newTask)
}

// List displays tasks with optional filtering
func (tm *TaskManager) List(ctx context.Context, showCompleted bool, filterPriority *task.Priority, searchTerm string, showDue string) error {
	tasks, err := tm.storage.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	// Filter tasks
	filtered := make([]*task.Task, 0)
	for _, task := range tasks {
		if !showCompleted && task.Completed {
			continue
		}
		if filterPriority != nil && task.Priority != *filterPriority {
			continue
		}
		if searchTerm != "" && !strings.Contains(strings.ToLower(task.Title), strings.ToLower(searchTerm)) &&
			!strings.Contains(strings.ToLower(task.Description), strings.ToLower(searchTerm)) {
			continue
		}
		if showDue != "" {
			dueFilter, err := filter.CreateTaskDueFilter(showDue)
			if err != nil {
				return fmt.Errorf("invalid due filter: %w", err)
			}
			if !dueFilter.Matches(task.DueDate) {
				continue
			}
		}
		filtered = append(filtered, task)
	}

	if len(filtered) == 0 {
		fmt.Println("No tasks match the current filters.")
		return nil
	}

	// Sort by priority (High -> Medium -> Low) and then by due date
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority > filtered[j].Priority // Higher priority first
		}
		if filtered[i].DueDate.IsZero() && !filtered[j].DueDate.IsZero() {
			return false
		}
		if !filtered[i].DueDate.IsZero() && filtered[j].DueDate.IsZero() {
			return true
		}
		return filtered[i].DueDate.Before(filtered[j].DueDate)
	})

	// Display tasks
	fmt.Printf("\n📋 Task List (%d tasks)\n", len(filtered))
	fmt.Println(strings.Repeat("=", 50))

	for _, t := range filtered {
		tm.displayTask(t)
		fmt.Println()
	}

	return nil
}

// Complete marks a task as completed
func (tm *TaskManager) Complete(ctx context.Context, id string) error {
	t, err := tm.storage.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	t.Complete()
	return tm.storage.Update(ctx, id, t)
}

// Uncomplete marks a task as not completed
func (tm *TaskManager) Uncomplete(ctx context.Context, id string) error {
	t, err := tm.storage.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	t.Uncomplete()
	return tm.storage.Update(ctx, id, t)
}

// Delete removes a task
func (tm *TaskManager) Delete(ctx context.Context, id string) error {
	// Check if task exists first
	_, err := tm.storage.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	return tm.storage.Delete(ctx, id)
}

// Update modifies an existing task
func (tm *TaskManager) Update(ctx context.Context, id, title, description string, priority task.Priority, dueDate time.Time) error {
	t, err := tm.storage.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if err := t.Update(title, description, priority, dueDate); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return tm.storage.Update(ctx, id, t)
}

// Show displays a single task by ID
func (tm *TaskManager) Show(ctx context.Context, id string) error {
	t, err := tm.storage.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	fmt.Printf("\n📝 Task Details\n")
	fmt.Println(strings.Repeat("=", 30))
	tm.displayTask(t)
	fmt.Println()

	return nil
}

// Stats displays task statistics
func (tm *TaskManager) Stats(ctx context.Context) error {
	tasks, err := tm.storage.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	var total, completed, overdue, dueToday, dueSoon int
	priorityCount := make(map[task.Priority]int)

	for _, t := range tasks {
		total++
		if t.Completed {
			completed++
		} else {
			if t.IsOverdue() {
				overdue++
			}
			if t.IsDueToday() {
				dueToday++
			}
			if t.IsDueSoon() {
				dueSoon++
			}
		}
		priorityCount[t.Priority]++
	}

	fmt.Printf("\n📊 Task Statistics\n")
	fmt.Println(strings.Repeat("=", 25))
	fmt.Printf("Total tasks: %d\n", total)
	fmt.Printf("Completed: %d\n", completed)
	fmt.Printf("Remaining: %d\n", total-completed)
	fmt.Printf("Overdue: %d\n", overdue)
	fmt.Printf("Due today: %d\n", dueToday)
	fmt.Printf("Due soon (7 days): %d\n", dueSoon)
	fmt.Println()
	fmt.Println("By Priority:")
	for p := task.High; p >= task.Low; p-- {
		fmt.Printf("  %s: %d\n", p.String(), priorityCount[p])
	}
	fmt.Println()

	return nil
}

// ExportTasks exports tasks to different formats
func (tm *TaskManager) ExportTasks(ctx context.Context, format string, filename string) error {
	tasks, err := tm.storage.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	switch strings.ToLower(format) {
	case "json":
		return tm.exportJSON(tasks, filename)
	case "csv":
		return tm.exportCSV(tasks, filename)
	case "markdown", "md":
		return tm.exportMarkdown(tasks, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// ConcurrentExport exports tasks to multiple formats concurrently
func (tm *TaskManager) ConcurrentExport(ctx context.Context, formats []string, baseFilename string) error {
	if len(formats) == 0 {
		return fmt.Errorf("no formats specified")
	}

	tasks, err := tm.storage.Load(ctx)
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
		go func(formatName string) {
			filename := baseFilename + "." + formatName
			var err error
			switch strings.ToLower(formatName) {
			case "json":
				err = tm.exportJSON(tasks, filename)
			case "csv":
				err = tm.exportCSV(tasks, filename)
			case "markdown", "md":
				err = tm.exportMarkdown(tasks, filename)
			default:
				err = fmt.Errorf("unsupported export format: %s", formatName)
			}
			results <- exportResult{format: formatName, err: err}
		}(format)
	}

	// Collect results
	var errors []string
	for i := 0; i < len(formats); i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.format, result.err))
		} else {
			fmt.Printf("✅ Exported to %s.%s\n", baseFilename, result.format)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("export errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// displayTask displays a single task in a formatted way
func (tm *TaskManager) displayTask(t *task.Task) {
	// Status icon and title
	status := "⏳"
	if t.Completed {
		status = "✅"
	} else if t.IsOverdue() {
		status = "🚨"
	} else if t.IsDueToday() {
		status = "📅"
	} else if t.IsDueSoon() {
		status = "⏰"
	}

	// Priority indicator
	priorityIcon := ""
	switch t.Priority {
	case task.High:
		priorityIcon = "🔴"
	case task.Medium:
		priorityIcon = "🟡"
	case task.Low:
		priorityIcon = "🟢"
	}

	fmt.Printf("%s %s %s\n", status, priorityIcon, t.Title)

	if t.Description != "" {
		fmt.Printf("   📝 %s\n", t.Description)
	}

	// Due date
	if !t.DueDate.IsZero() {
		dueStr := t.DueDate.Format("2006-01-02 15:04")
		if t.IsOverdue() {
			fmt.Printf("   ⏰ Due: %s (OVERDUE)\n", dueStr)
		} else if t.IsDueToday() {
			fmt.Printf("   ⏰ Due: %s (TODAY)\n", dueStr)
		} else {
			fmt.Printf("   ⏰ Due: %s\n", dueStr)
		}
	}

	// ID and timestamps
	fmt.Printf("   🆔 ID: %s\n", t.ID)
	fmt.Printf("   📅 Created: %s\n", t.CreatedAt.Format("2006-01-02 15:04"))
	if t.UpdatedAt.After(t.CreatedAt) {
		fmt.Printf("   🔄 Updated: %s\n", t.UpdatedAt.Format("2006-01-02 15:04"))
	}
}

// exportJSON exports tasks to JSON format
func (tm *TaskManager) exportJSON(tasks []*task.Task, filename string) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// exportCSV exports tasks to CSV format
func (tm *TaskManager) exportCSV(tasks []*task.Task, filename string) error {
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
			strings.ReplaceAll(t.Title, ",", ";"), // Escape commas
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

// exportMarkdown exports tasks to Markdown format
func (tm *TaskManager) exportMarkdown(tasks []*task.Task, filename string) error {
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
			tm.writeMarkdownTask(file, t)
		}
		fmt.Fprintln(file)
	}

	// Write completed tasks
	if len(completed) > 0 {
		fmt.Fprintf(file, "## Completed Tasks (%d)\n\n", len(completed))
		for _, t := range completed {
			tm.writeMarkdownTask(file, t)
		}
	}

	return nil
}

// writeMarkdownTask writes a single task in Markdown format
func (tm *TaskManager) writeMarkdownTask(file *os.File, t *task.Task) {
	// Task header
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

	// Description
	if t.Description != "" {
		fmt.Fprintf(file, "**Description:** %s\n\n", t.Description)
	}

	// Due date
	if !t.DueDate.IsZero() {
		fmt.Fprintf(file, "**Due:** %s\n\n", t.DueDate.Format("2006-01-02 15:04"))
	}

	// Metadata
	fmt.Fprintf(file, "**ID:** `%s`  \n", t.ID)
	fmt.Fprintf(file, "**Created:** %s  \n", t.CreatedAt.Format("2006-01-02 15:04"))
	if t.UpdatedAt.After(t.CreatedAt) {
		fmt.Fprintf(file, "**Updated:** %s  \n", t.UpdatedAt.Format("2006-01-02 15:04"))
	}
	fmt.Fprintln(file)
}
