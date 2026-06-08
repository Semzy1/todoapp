# todoapp

A simple command-line todo application with optional web interface, written in Go.

## Features

- **CLI Mode**: Interactive command-line interface with commands for managing tasks
- **Web Mode**: HTTP server for managing tasks through a web browser
- **Task Management**: Add, list, update, complete, and delete tasks
- **Task Filtering**: List tasks by status (all, pending, done)

## Commands

When running in CLI mode, the following commands are available:

| Command | Description |
|---------|-------------|
| `add <task>` | Add a new task |
| `list [all|pending|done]` | Show tasks (defaults to all) |
| `done <id>` | Mark task as completed |
| `delete <id>` | Remove a task |
| `edit <id> <text>` | Update task title |
| `web` | Start the web server |
| `help` | Show help information |
| `quit` | Exit the application |

## Usage

```bash
# Run in CLI mode
go run .

# Run in web mode (default port 8080)
go run . --web

# Run web mode on a specific port
PORT=3000 go run . --web
```

## Running Tests

```bash
go test -v ./...
```

## Project Structure

- `main.go` - CLI and entry point
- `todo.go` - Task and TodoList data structures
- `web.go` - Web server implementation
- `index.html` - Main web page template
- `detail.html` - Task detail page template
