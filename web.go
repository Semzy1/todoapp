package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var listenAndServe = http.ListenAndServe

type WebServer struct {
	todo *TodoList
}

func NewWebServer(todo *TodoList) *WebServer {
	return &WebServer{todo: todo}
}

func (ws *WebServer) Start(port string) error {
	// Parse HTML templates from files in the project root
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Warning: index.html not found, please ensure it exists")
		return err
	}

	detailTmpl, err := template.ParseFiles("detail.html")
	if err != nil {
		log.Printf("Warning: detail.html not found, please ensure it exists")
		return err
	}

	// Routes
	// Serve the detail template directly (renders with no task -> shows 'Task Not Found')
	http.HandleFunc("/detail.html", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Task              *Task
			TaskJSON          string
			TimeSinceCreation string
			TimeToComplete    string
		}{
			Task: nil,
		}
		if err := detailTmpl.Execute(w, data); err != nil {
			log.Printf("❌ Error executing detail template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/", ws.handleIndex(tmpl))
	http.HandleFunc("/task/", ws.handleTaskDetail(detailTmpl))
	http.HandleFunc("/add", ws.handleAdd)
	http.HandleFunc("/complete", ws.handleComplete)
	http.HandleFunc("/delete", ws.handleDelete)
	http.HandleFunc("/edit", ws.handleEdit)
	http.HandleFunc("/api/tasks", ws.handleAPI)
	http.HandleFunc("/export/json", ws.handleExport)

	log.Printf("========================================")
	log.Printf("✅ Web server starting on http://localhost%s", port)
	log.Printf("📊 Watch real-time activity in this terminal")
	log.Printf("========================================")
	return listenAndServe(port, nil)
}

func (ws *WebServer) handleIndex(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🌐 Home page accessed from %s", r.RemoteAddr)

		filter := r.URL.Query().Get("filter")
		if filter == "" {
			filter = "all"
		}

		var tasks []Task

		switch filter {
		case "pending":
			tasks = ws.todo.GetByStatus(false)
			log.Printf("🔍 Filtering: PENDING tasks")
		case "completed":
			tasks = ws.todo.GetByStatus(true)
			log.Printf("🔍 Filtering: COMPLETED tasks")
		default:
			tasks = ws.todo.GetAll()
			log.Printf("🔍 Filtering: ALL tasks")
		}

		stats := struct {
			Total     int
			Pending   int
			Completed int
		}{
			Total:     len(ws.todo.GetAll()),
			Pending:   len(ws.todo.GetByStatus(false)),
			Completed: len(ws.todo.GetByStatus(true)),
		}

		data := struct {
			Tasks []Task
			Stats struct {
				Total     int
				Pending   int
				Completed int
			}
			Filter string
		}{
			Tasks:  tasks,
			Stats:  stats,
			Filter: filter,
		}

		log.Printf("📊 Current stats: %d total, %d pending, %d completed",
			stats.Total, stats.Pending, stats.Completed)

		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("❌ Error executing template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (ws *WebServer) handleTaskDetail(detailTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path /task/123
		idStr := r.URL.Path[len("/task/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("❌ Invalid task ID: %s", idStr)
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		log.Printf("🔍 Viewing task details for ID: %d from %s", id, r.RemoteAddr)

		task, err := ws.todo.GetTaskByID(id)
		if err != nil {
			log.Printf("❌ Task %d not found", id)
			// Render detail page with error message
			data := struct {
				Task              *Task
				TaskJSON          string
				TimeSinceCreation string
				TimeToComplete    string
			}{
				Task: nil,
			}
			detailTmpl.Execute(w, data)
			return
		}

		// Calculate time since creation
		now := time.Now()
		duration := now.Sub(task.CreatedAt)
		hours := int(duration.Hours())
		days := hours / 24
		hours = hours % 24
		minutes := int(duration.Minutes()) % 60

		timeSinceCreation := ""
		if days > 0 {
			timeSinceCreation = fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
		} else if hours > 0 {
			timeSinceCreation = fmt.Sprintf("%d hours, %d minutes", hours, minutes)
		} else {
			timeSinceCreation = fmt.Sprintf("%d minutes", minutes)
		}

		// Calculate time to complete if done
		timeToComplete := ""
		if task.Done {
			durationToComplete := task.UpdatedAt.Sub(task.CreatedAt)
			hoursToComplete := int(durationToComplete.Hours())
			daysToComplete := hoursToComplete / 24
			hoursToComplete = hoursToComplete % 24
			minutesToComplete := int(durationToComplete.Minutes()) % 60

			if daysToComplete > 0 {
				timeToComplete = fmt.Sprintf("%d days, %d hours, %d minutes", daysToComplete, hoursToComplete, minutesToComplete)
			} else if hoursToComplete > 0 {
				timeToComplete = fmt.Sprintf("%d hours, %d minutes", hoursToComplete, minutesToComplete)
			} else {
				timeToComplete = fmt.Sprintf("%d minutes", minutesToComplete)
			}
		}

		// Get JSON representation of the task
		jsonBytes, _ := json.MarshalIndent(task, "", "  ")
		taskJSON := string(jsonBytes)

		data := struct {
			Task              *Task
			TaskJSON          string
			TimeSinceCreation string
			TimeToComplete    string
		}{
			Task:              task,
			TaskJSON:          taskJSON,
			TimeSinceCreation: timeSinceCreation,
			TimeToComplete:    timeToComplete,
		}

		log.Printf("📄 Displaying task: #%d - '%s' (Done: %v)", task.ID, task.Title, task.Done)

		err = detailTmpl.Execute(w, data)
		if err != nil {
			log.Printf("❌ Error executing detail template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (ws *WebServer) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("❌ Invalid method for /add: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		log.Printf("❌ Add task failed: empty title from %s", r.RemoteAddr)
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	id, err := ws.todo.Add(title)
	if err != nil {
		log.Printf("❌ Add task failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("✅✅✅ TASK ADDED: ID=%d, Title='%s' from %s", id, title, r.RemoteAddr)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ws *WebServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("❌ Invalid method for /complete: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("❌ Complete task failed: invalid ID '%s'", idStr)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = ws.todo.Complete(id)
	if err != nil {
		log.Printf("❌ Complete task failed: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("✅✅✅ TASK COMPLETED: ID=%d from %s", id, r.RemoteAddr)

	// Check referer to redirect back to detail page if needed
	referer := r.Header.Get("Referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (ws *WebServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("❌ Invalid method for /delete: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("❌ Delete task failed: invalid ID '%s'", idStr)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = ws.todo.Remove(id)
	if err != nil {
		log.Printf("❌ Delete task failed: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("🗑️🗑️🗑️ TASK DELETED: ID=%d from %s", id, r.RemoteAddr)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ws *WebServer) handleEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("❌ Invalid method for /edit: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("❌ Edit task failed: invalid ID '%s'", idStr)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		log.Printf("❌ Edit task failed: empty title for ID=%d", id)
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	oldTitle := ""
	if task, err := ws.todo.GetTaskByID(id); err == nil {
		oldTitle = task.Title
	}

	err = ws.todo.Update(id, title)
	if err != nil {
		log.Printf("❌ Edit task failed: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("✏️✏️✏️ TASK EDITED: ID=%d, Old='%s' → New='%s' from %s", id, oldTitle, title, r.RemoteAddr)

	// Check referer to redirect back to detail page if needed
	referer := r.Header.Get("Referer")
	if referer != "" && referer != "http://localhost"+r.Host+"/" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (ws *WebServer) handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("🌐 API request: %s /api/tasks from %s", r.Method, r.RemoteAddr)

	var response interface{}

	switch r.Method {
	case http.MethodGet:
		response = ws.todo.GetAll()
		log.Printf("📤 Returning %d tasks via API", len(ws.todo.GetAll()))
	case http.MethodPost:
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.Printf("❌ API POST failed: invalid JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		id, err := ws.todo.Add(task.Title)
		if err != nil {
			log.Printf("❌ API POST failed: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response = map[string]int{"id": id}
		log.Printf("📥 API POST: Added task '%s' with ID=%d", task.Title, id)
	default:
		log.Printf("❌ API method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ API encode failed: %v", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func (ws *WebServer) handleExport(w http.ResponseWriter, r *http.Request) {
	jsonData, err := ws.todo.ToJSON()
	if err != nil {
		log.Printf("❌ Export failed: %v", err)
		http.Error(w, "Failed to generate JSON", http.StatusInternalServerError)
		return
	}

	taskCount := len(ws.todo.GetAll())
	log.Printf("📥📥📥 JSON EXPORT: Downloading %d tasks (%d bytes) from %s", taskCount, len(jsonData), r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=todo_export_"+time.Now().Format("20060102_150405")+".json")
	w.Write(jsonData)
}
