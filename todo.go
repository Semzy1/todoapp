package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Task - yeah just a simple struct for each todo item
// Added UpdatedAt because why not, helps with debugging
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"` // "Completed" felt too long, Done is fine
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// The main struct - holds tasks in memory
type TodoList struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

// Constructor - Go doesn't have constructors so just a helper
func NewTodoList() *TodoList {
	return &TodoList{
		Tasks:  []Task{},
		NextID: 1,
	}
}

// Add - creates new task
// Returns error if title is empty (user probably fat-fingered)
func (t *TodoList) Add(title string) (int, error) {
	// Trim spaces - people sometimes add spaces by accident
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, errors.New("title cannot be empty, try again")
	}

	now := time.Now()
	task := Task{
		ID:        t.NextID,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Tasks = append(t.Tasks, task)
	t.NextID++
	return task.ID, nil
}

func (t *TodoList) GetAll() []Task {
	return t.Tasks
}

// GetByStatus - filter by done/pending
// I could've made this more generic but this is fine
func (t *TodoList) GetByStatus(done bool) []Task {
	result := make([]Task, 0)
	for _, task := range t.Tasks {
		if task.Done == done {
			result = append(result, task)
		}
	}
	return result
}

// Complete - marks task as done
// Returns error if ID not found - user might typo
func (t *TodoList) Complete(id int) error {
	for i := range t.Tasks {
		if t.Tasks[i].ID == id {
			// Already done? Should we error or just ignore?
			// I'll just ignore and update timestamp anyway
			t.Tasks[i].Done = true
			t.Tasks[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("task with ID %d doesn't exist", id)
}

// Remove - deletes task by ID
// O(n) operation but with small lists who cares
func (t *TodoList) Remove(id int) error {
	for i, task := range t.Tasks {
		if task.ID == id {
			// Slice trick to remove element
			// This is standard Go pattern
			t.Tasks = append(t.Tasks[:i], t.Tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("can't find task %d", id)
}

// Update - changes task title
func (t *TodoList) Update(id int, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("title can't be empty")
	}

	for i := range t.Tasks {
		if t.Tasks[i].ID == id {
			t.Tasks[i].Title = title
			t.Tasks[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("task %d not in list", id)
}

// ToJSON - export tasks as JSON (for the JSON requirement)
func (t *TodoList) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// GetTaskByID - helper for the web UI
func (t *TodoList) GetTaskByID(id int) (*Task, error) {
	for i := range t.Tasks {
		if t.Tasks[i].ID == id {
			return &t.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task %d not found", id)
}
