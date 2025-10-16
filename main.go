package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-fun/internal/cli"
	"go-fun/internal/storage"
	"go-fun/internal/task"
)

const (
	appName        = "go-fun"
	appVersion     = "1.0.0"
	appDescription = "A simple CLI task manager built with Go"
)

var (
	version = flag.Bool("version", false, "Show version information")
	help    = flag.Bool("help", false, "Show help information")
	dataDir = flag.String("data-dir", "", "Directory to store task data (default: ~/.go-fun)")
)

func main() {
	// Parse global flags
	flag.Parse()
	args := flag.Args()

	if *version {
		showVersion()
		return
	}

	if *help || len(args) == 0 {
		showHelp()
		return
	}

	// Set up data directory
	dataPath := getDataPath()

	// Initialize storage
	jsonStorage := storage.NewJSONFileStorage(filepath.Join(dataPath, "tasks.json"))

	// Create task manager
	taskManager := cli.NewTaskManager(jsonStorage)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command
	command := args[0]
	commandArgs := args[1:]

	if err := executeCommand(ctx, taskManager, command, commandArgs); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func executeCommand(ctx context.Context, tm *cli.TaskManager, command string, args []string) error {
	switch command {
	case "add":
		return handleAdd(ctx, tm, args)
	case "list", "ls":
		return handleList(ctx, tm, args)
	case "complete", "done":
		return handleComplete(ctx, tm, args)
	case "uncomplete", "undo":
		return handleUncomplete(ctx, tm, args)
	case "delete", "rm":
		return handleDelete(ctx, tm, args)
	case "update", "edit":
		return handleUpdate(ctx, tm, args)
	case "show", "get":
		return handleShow(ctx, tm, args)
	case "stats":
		return handleStats(ctx, tm, args)
	case "export":
		return handleExport(ctx, tm, args)
	case "export-all":
		return handleExportAll(ctx, tm, args)
	case "watch":
		return handleWatch(ctx, tm, args)
	default:
		return fmt.Errorf("unknown command: %s. Use 'go-fun -help' for usage", command)
	}
}

func normalizeTags(in []string) []string {
	set := make(map[string]struct{}, len(in))
	for _, v := range in {
		set[v] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func handleAdd(ctx context.Context, tm *cli.TaskManager, args []string) error {
	flagSet := flag.NewFlagSet("add", flag.ContinueOnError)

	title := ""
	description := ""
	dueDateStr := ""
	priorityStr := ""

	dueDate := time.Time{}
	priority := task.Medium
	tags := make(cli.TagList, 0)

	titleDesc := "Title for the task"
	flagSet.StringVar(&title, "t", title, titleDesc)
	flagSet.StringVar(&title, "title", title, titleDesc)

	descDesc := "Description for the task"
	flagSet.StringVar(&description, "d", description, descDesc)
	flagSet.StringVar(&description, "desc", description, descDesc)
	flagSet.StringVar(&description, "description", description, descDesc)

	duedateDesc := "Due date for the task (yyyy-mm-dd)"
	flagSet.StringVar(&dueDateStr, "D", dueDateStr, duedateDesc)
	flagSet.StringVar(&dueDateStr, "duedate", dueDateStr, duedateDesc)

	priorityDesc := "Priority for the task (l, m, h)"
	flagSet.StringVar(&priorityStr, "p", priorityStr, priorityDesc)
	flagSet.StringVar(&priorityStr, "priority", priorityStr, priorityDesc)

	tagDesc := "Tag for the task (repeatable or comma-separated)"
	flagSet.Var(&tags, "T", tagDesc)
	flagSet.Var(&tags, "tag", tagDesc)

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	// -t --title
	if title == "" {
		return fmt.Errorf("title is required")
	}
	// -d --desc --description
	if description == "" {
		return fmt.Errorf("description is required")
	}
	// -p --priority
	if priorityStr != "" {
		switch strings.ToLower(priorityStr) {
		case "low", "l":
			priority = task.Low
		case "medium", "med", "m":
			priority = task.Medium
		case "high", "h":
			priority = task.High
		default:
			return fmt.Errorf("invalid priority: %s. Use: low, medium, high", priorityStr)
		}
	}
	// -D --duedate
	if dueDateStr != "" {
		parsedDate, err := parseDate(dueDateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		dueDate = parsedDate
	}

	// -T --tag
	normalizedTags := normalizeTags(tags)

	return tm.Add(ctx, title, description, priority, dueDate, normalizedTags)
}

func handleList(ctx context.Context, tm *cli.TaskManager, args []string) error {
	var filterPriority *task.Priority
	searchTerm := ""
	showCompleted := false
	showDue := ""

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "-c", "--completed":
			showCompleted = true
		case "-d", "--due":
			if i+1 < len(args) {
				showDue = args[i+1]
			}
		case "-p", "--priority":
			if i+1 < len(args) {
				switch strings.ToLower(args[i+1]) {
				case "low", "l":
					p := task.Low
					filterPriority = &p
				case "medium", "med", "m":
					p := task.Medium
					filterPriority = &p
				case "high", "h":
					p := task.High
					filterPriority = &p
				}
			}
		case "-s", "--search":
			if i+1 < len(args) {
				searchTerm = args[i+1]
			}
		}
	}

	return tm.List(ctx, showCompleted, filterPriority, searchTerm, showDue)
}

func handleComplete(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: complete <task-id>")
	}

	return tm.Complete(ctx, args[0])
}

func handleUncomplete(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: uncomplete <task-id>")
	}

	return tm.Uncomplete(ctx, args[0])
}

func handleDelete(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: delete <task-id>")
	}

	return tm.Delete(ctx, args[0])
}

func handleUpdate(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: update <task-id> <title> [description] [priority] [due-date]")
	}

	id := args[0]
	title := args[1]
	description := ""
	priority := task.Medium
	dueDate := time.Time{}

	if len(args) > 2 {
		description = args[2]
	}
	if len(args) > 3 {
		switch strings.ToLower(args[3]) {
		case "low", "l":
			priority = task.Low
		case "medium", "med", "m":
			priority = task.Medium
		case "high", "h":
			priority = task.High
		default:
			return fmt.Errorf("invalid priority: %s. Use: low, medium, high", args[3])
		}
	}
	if len(args) > 4 {
		parsedDate, err := parseDate(args[4])
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		dueDate = parsedDate
	}

	return tm.Update(ctx, id, title, description, priority, dueDate)
}

func handleShow(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: show <task-id>")
	}

	return tm.Show(ctx, args[0])
}

func handleStats(ctx context.Context, tm *cli.TaskManager, args []string) error {
	return tm.Stats(ctx)
}

func handleExport(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: export <format> <filename>")
	}

	format := args[0]
	filename := args[1]

	return tm.ExportTasks(ctx, format, filename)
}

func handleExportAll(ctx context.Context, tm *cli.TaskManager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: export-all <formats> <base-filename>")
	}

	// Parse formats (comma-separated)
	formatsStr := args[0]
	baseFilename := args[1]

	formats := strings.Split(formatsStr, ",")
	for i, format := range formats {
		formats[i] = strings.TrimSpace(format)
	}

	fmt.Printf("🚀 Starting concurrent export to %d formats...\n", len(formats))
	return tm.ConcurrentExport(ctx, formats, baseFilename)
}

func handleWatch(ctx context.Context, tm *cli.TaskManager, args []string) error {
	// todo: playground area...
	return nil
}

func parseDate(dateStr string) (time.Time, error) {
	// Handle special cases first
	switch strings.ToLower(dateStr) {
	case "today":
		return time.Now().Truncate(24 * time.Hour), nil
	case "tomorrow":
		return time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour), nil
	}

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"01/02/2006 15:04",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// Try parsing as duration (e.g., "1d", "2h", "30m")
	// Handle "d" suffix for days
	if strings.HasSuffix(dateStr, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(dateStr, "d")); err == nil {
			return time.Now().Add(time.Duration(days) * 24 * time.Hour), nil
		}
	}

	// Try standard duration parsing
	if duration, err := time.ParseDuration(dateStr); err == nil {
		return time.Now().Add(duration), nil
	}

	// Try parsing as days from now (e.g., "3" means 3 days from now)
	if days, err := strconv.Atoi(dateStr); err == nil {
		return time.Now().Add(time.Duration(days) * 24 * time.Hour), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func getDataPath() string {
	if *dataDir != "" {
		return *dataDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dataPath := filepath.Join(homeDir, ".go-fun")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	return dataPath
}

func showVersion() {
	fmt.Printf("%s version %s\n", appName, appVersion)
	fmt.Printf("Description: %s\n", appDescription)
}

func showHelp() {
	fmt.Printf("%s - %s\n\n", appName, appDescription)
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-flags] <command> [command-flags] [arguments]\n\n", appName)

	fmt.Println("Global Flags:")
	fmt.Println("  -version     Show version information")
	fmt.Println("  -help        Show this help message")
	fmt.Println("  -data-dir    Directory to store task data (default: ~/.go-fun)")
	fmt.Println()

	fmt.Println("Commands:")
	fmt.Println("  add [-t --title ...] [-d --desc --description ...] [-p --priority ...] [-D --duedate ...] [-T --tag ...]")
	fmt.Println("    Add a new task")
	fmt.Println("    Priority: l/low, m/med/medium, h/high (default: medium)")
	fmt.Println("    Duedate formats: 2006-01-02, 01/02/2006, tomorrow, 1d, 3")
	fmt.Println("    Tag: Single, repeated flag, or comma-separated strings")
	fmt.Println()

	fmt.Println("  list [flags]")
	fmt.Println("    List tasks")
	fmt.Println("    Flags:")
	fmt.Println("      -c, --completed    Show completed tasks")
	fmt.Println("      -d, --due          Filter by due date (today, overdue, week, 3)")
	fmt.Println("      -p, --priority     Filter by priority (low/medium/high)")
	fmt.Println("      -s, --search       Search in title and description")
	fmt.Println()

	fmt.Println("  complete <task-id>")
	fmt.Println("    Mark a task as completed")
	fmt.Println()

	fmt.Println("  uncomplete <task-id>")
	fmt.Println("    Mark a task as not completed")
	fmt.Println()

	fmt.Println("  delete <task-id>")
	fmt.Println("    Delete a task")
	fmt.Println()

	fmt.Println("  update <task-id> <title> [description] [priority] [due-date]")
	fmt.Println("    Update an existing task")
	fmt.Println()

	fmt.Println("  show <task-id>")
	fmt.Println("    Show details of a specific task")
	fmt.Println()

	fmt.Println("  stats")
	fmt.Println("    Show task statistics")
	fmt.Println()

	fmt.Println("  export <format> <filename>")
	fmt.Println("    Export tasks to file")
	fmt.Println("    Formats: json, csv, markdown")
	fmt.Println()

	fmt.Println("  export-all <formats> <base-filename>")
	fmt.Println("    Export tasks to multiple formats concurrently")
	fmt.Println("    Formats: comma-separated list (e.g., json,csv,markdown)")
	fmt.Println()

	fmt.Println("Examples:")
	fmt.Printf("  %s add \"Learn Go\" \"Complete Go tutorial\" high tomorrow\n", appName)
	fmt.Printf("  %s list\n", appName)
	fmt.Printf("  %s list -p high -s learn\n", appName)
	fmt.Printf("  %s complete task_1234567890\n", appName)
	fmt.Printf("  %s export json tasks.json\n", appName)
	fmt.Printf("  %s export-all json,csv,markdown backup\n", appName)
}
