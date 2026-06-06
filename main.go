package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	todo := NewTodoList()

	// Check if we should run web server or CLI
	if len(os.Args) > 1 && os.Args[1] == "--web" {
		// Run web server
		port := ":8080"
		if len(os.Args) > 2 {
			port = ":" + os.Args[2]
		}

		server := NewWebServer(todo)
		if err := server.Start(port); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run CLI version
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Simple Todo App")
	fmt.Println("Type help for commands")
	fmt.Println("Tip: Run with '--web' flag to start web server")

	for {
		fmt.Print("todo> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if input == "" {
					fmt.Println()
					return
				}
				input = strings.TrimSpace(input)
			} else {
				fmt.Fprintln(os.Stderr, "Error reading input:", err)
				continue
			}
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit", "q":
			fmt.Println("Goodbye!")
			return

		case "help", "h", "?":
			printHelp()

		case "add", "a":
			if len(parts) < 2 {
				fmt.Println("Usage: add <task description>")
				continue
			}
			title := strings.Join(parts[1:], " ")
			id, err := todo.Add(title)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("Added task %d\n", id)
			}

		case "list", "ls":
			if len(parts) == 1 {
				printTasks(todo.GetAll())
				continue
			}

			switch parts[1] {
			case "all", "a":
				printTasks(todo.GetAll())
			case "pending", "p":
				printTasks(todo.GetByStatus(false))
			case "done", "d":
				printTasks(todo.GetByStatus(true))
			default:
				fmt.Println("Usage: list [all|pending|done]")
			}

		case "done", "complete", "finish":
			if len(parts) < 2 {
				fmt.Println("Usage: done <task-id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("ID must be a number")
				continue
			}
			if err := todo.Complete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Task done!")
			}

		case "rm", "delete", "del":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <task-id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("ID must be a number")
				continue
			}
			if err := todo.Remove(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Task removed")
			}

		case "edit", "update":
			if len(parts) < 3 {
				fmt.Println("Usage: edit <id> <new title>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("ID must be a number")
				continue
			}
			newTitle := strings.Join(parts[2:], " ")
			if err := todo.Update(id, newTitle); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Task updated")
			}

		case "web":
			fmt.Println("Starting web server on http://localhost:8080")
			fmt.Println("Press Ctrl+C to stop the server")
			server := NewWebServer(todo)
			if err := server.Start(":8080"); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			}
			return

		default:
			fmt.Println("Unknown command. Type help.")
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  add <task>        - Add a task")
	fmt.Println("  list [all|pending|done] - Show tasks")
	fmt.Println("  done <id>         - Mark task done")
	fmt.Println("  delete <id>       - Remove task")
	fmt.Println("  edit <id> <text>  - Change task title")
	fmt.Println("  web               - Start web server")
	fmt.Println("  help              - Show this help")
	fmt.Println("  quit              - Exit")
}

func printTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks yet. Use add <task>.")
		return
	}

	for _, task := range tasks {
		status := " "
		if task.Done {
			status = "x"
		}
		fmt.Printf("%d. [%s] %s (created: %s)\n", task.ID, status, task.Title, task.CreatedAt.Format("15:04"))
	}
}

