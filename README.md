# TodoApp

A simple Todo application built with Go that supports both a Command-Line Interface (CLI) and a Web Interface. The application allows users to create, manage, update, and track tasks efficiently while demonstrating core Go concepts such as structs, methods, HTTP servers, templates, JSON handling, and testing.

---

## Features

### CLI Mode
- Add new tasks
- List all tasks
- Filter tasks by status
- Edit task titles
- Mark tasks as completed
- Delete tasks
- Interactive command-based interface

### Web Mode
- View tasks in a browser
- Create new tasks through a web form
- View task details
- Edit existing tasks
- Mark tasks as completed
- Delete tasks

### Task Management
- Create tasks
- Update task details
- Track completion status
- Delete tasks
- Filter tasks by:
  - All
  - Pending
  - Completed

---

## Project Structure

```text
todoapp/
│
├── main.go         # Application entry point and CLI logic
├── todo.go         # Task and TodoList data structures
├── web.go          # HTTP server and web handlers
├── index.html      # Main task list page
├── detail.html     # Task detail page
├── go.mod          # Go module definition
├── go.sum          # Dependency checksums
└── README.md       # Project documentation
```

---

## Installation

### Clone the Repository

```bash
git clone https://github.com/your-username/todoapp.git
cd todoapp
```

### Install Dependencies

```bash
go mod tidy
```

---

## Running the Application

### Run in CLI Mode

```bash
go run .
```

### Run in Web Mode

```bash
go run . --web
```

### Run Web Mode on a Custom Port

Linux/macOS:

```bash
PORT=3000 go run . --web
```

Windows (PowerShell):

```powershell
$env:PORT="3000"
go run . --web
```

After starting the web server, open:

```text
http://localhost:3000
```

or

```text
http://localhost:8080
```

depending on the configured port.

---

## CLI Commands

| Command | Description |
|----------|-------------|
| `add <task>` | Add a new task |
| `list` | List all tasks |
| `list pending` | Show pending tasks |
| `list done` | Show completed tasks |
| `done <id>` | Mark a task as completed |
| `edit <id> <text>` | Update task title |
| `delete <id>` | Delete a task |
| `web` | Start the web server |
| `help` | Show available commands |
| `quit` | Exit the application |

---

## Example Usage

### Add a Task

```text
> add Buy groceries
Task added successfully
```

### View Tasks

```text
> list

1. [ ] Buy groceries
2. [x] Complete assignment
```

### Complete a Task

```text
> done 1
Task marked as completed
```

---

## Running Tests

Execute all tests with:

```bash
go test -v ./...
```

---

## Technologies Used

- Go (Golang)
- HTML Templates
- HTTP Server
- JSON Encoding/Decoding
- Go Testing Package

---

## Learning Objectives

This project demonstrates:

- Structs and methods
- Package organization
- Error handling
- JSON processing
- HTTP server development
- HTML template rendering
- CRUD operations
- Unit testing
- Command-line application development

---

## Future Improvements

- Persistent storage using SQLite or PostgreSQL
- User authentication
- Task categories
- Due dates and reminders
- REST API endpoints
- Docker support
- Deployment to cloud platforms

---

## License

This project is for educational purposes and learning Go application development.
