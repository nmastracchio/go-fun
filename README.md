# Go-Fun: A CLI Task Manager

A feature-rich command-line task manager built with Go that demonstrates core Go concepts including structs, interfaces, concurrency, error handling, and testing.

## Features

### Core Functionality
- ✅ Add, list, complete, delete, and update tasks
- ✅ Task priorities (low, medium, high) and due dates
- ✅ Filter and search tasks
- ✅ Persistent storage using JSON
- ✅ Colorful terminal output with emojis

### Advanced Features
- 🚀 **Concurrent exports** - Export to multiple formats simultaneously
- 📊 **Task statistics** - Overview of your productivity
- 🔍 **Flexible filtering** - Search by priority, completion status, or text
- 📅 **Smart date parsing** - Support for various date formats
- 🎨 **Beautiful output** - Color-coded priorities and status indicators

### Go Concepts Demonstrated
- **Structs & Methods**: Task type with validation and behavior methods
- **Interfaces**: Storage abstraction for different backends
- **Error Handling**: Custom error types and proper error wrapping
- **Concurrency**: Goroutines and channels for concurrent operations
- **File I/O**: JSON marshaling/unmarshaling with atomic writes
- **Testing**: Table-driven tests, benchmarks, and mock implementations
- **Package Organization**: Proper Go project layout with internal packages

## Installation

### Prerequisites
- Go 1.21 or later
- Git (for cloning)

### Build from Source
```bash
git clone <repository-url>
cd go-fun
go build -o go-fun .
```

## Usage

### Basic Commands

```bash
# Show help
go-fun -help

# Add a new task
go-fun add "Learn Go concurrency" "Study goroutines and channels" high tomorrow

# List all tasks
go-fun list

# List with filters
go-fun list -p high -s learn

# Complete a task
go-fun complete task_1234567890

# Update a task
go-fun update task_1234567890 "Updated title" "New description" medium 3d

# Delete a task
go-fun delete task_1234567890

# Show task statistics
go-fun stats
```

### Advanced Commands

```bash
# Export to single format
go-fun export json tasks.json
go-fun export csv tasks.csv
go-fun export markdown tasks.md

# Export to multiple formats concurrently
go-fun export-all json,csv,markdown backup
```

### Date Formats

The CLI supports various date formats:
- `2025-10-15` - ISO date
- `10/15/2025` - US date format
- `tomorrow` - Relative date
- `3d` - Days from now
- `2h` - Hours from now
- `1w` - Weeks from now

### Filtering Options

```bash
# Show completed tasks
go-fun list -c

# Filter by priority
go-fun list -p high
go-fun list -p medium
go-fun list -p low

# Search in title and description
go-fun list -s "learn go"
```

## Project Structure

```
go-fun/
├── main.go                     # Entry point and CLI parsing
├── go.mod                      # Go module definition
├── internal/                   # Internal packages
│   ├── task/
│   │   ├── task.go            # Task struct and methods
│   │   └── task_test.go       # Task tests and benchmarks
│   ├── storage/
│   │   ├── storage.go         # Storage interface and implementations
│   │   ├── storage_test.go    # Storage tests and benchmarks
│   │   └── concurrent_storage.go # Concurrency features
│   └── cli/
│       ├── commands.go        # CLI command implementations
│       └── commands_test.go   # CLI tests and benchmarks
└── README.md                  # This file
```

## Architecture

### Task Model
The `Task` struct represents a single todo item with:
- Unique ID generation
- Title and description with validation
- Priority levels (Low, Medium, High)
- Due date with smart parsing
- Completion status
- Timestamps for creation and updates

### Storage Layer
The storage system uses interfaces for flexibility:
- `Storage` interface defines operations
- `JSONFileStorage` for persistent file-based storage
- `InMemoryStorage` for testing and temporary storage
- `ConcurrentStorage` wrapper for background operations

### CLI Layer
The CLI provides a user-friendly interface:
- Command parsing with the `flag` package
- Rich output formatting with emojis
- Error handling with helpful messages
- Concurrent operations for performance

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Run specific package tests
go test ./internal/task/...
```

### Test Coverage
The project includes comprehensive test coverage:
- **Unit tests** for all core functionality
- **Integration tests** for storage operations
- **Benchmark tests** for performance measurement
- **Table-driven tests** for validation logic
- **Error handling tests** for edge cases

### Code Quality
- Follows Go best practices and conventions
- Uses proper error handling with wrapped errors
- Implements interfaces for testability
- Includes comprehensive documentation
- Uses context for cancellation support

## Examples

### Creating a Learning Plan
```bash
# Add learning tasks
go-fun add "Learn Go basics" "Complete Go tutorial" high today
go-fun add "Practice concurrency" "Build concurrent web scraper" medium tomorrow
go-fun add "Study interfaces" "Read Effective Go chapter" low 3d

# Check progress
go-fun stats

# Export for sharing
go-fun export-all json,csv,markdown learning-plan
```

### Daily Task Management
```bash
# Morning planning
go-fun list -p high

# Afternoon review
go-fun list -s "urgent"

# Evening cleanup
go-fun list -c
go-fun export json daily-report.json
```

## Performance

Benchmark results on typical hardware:
- Task creation: ~530ns per task
- Task validation: ~2.4ns per validation
- In-memory storage: ~208µs per operation
- File storage: ~4.5ms per operation
- Concurrent exports: 3x faster than sequential

## Contributing

This project serves as a learning example for Go development. Key areas for contribution:
- Additional storage backends (database, cloud)
- More export formats (XML, YAML)
- Enhanced filtering and sorting
- Web interface
- Plugin system

## License

This project is for educational purposes and demonstrates Go best practices.

## Learning Outcomes

By studying this codebase, you'll learn:
- Go project structure and organization
- Interface design and implementation
- Error handling patterns
- Concurrent programming with goroutines and channels
- Testing strategies and benchmarking
- CLI application development
- JSON marshaling and file I/O
- Package management and dependencies

Happy coding! 🚀
