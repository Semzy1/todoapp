package main

import (
	"testing"
	"time"
)

// Test adding tasks - basic stuff
func TestAddTask(t *testing.T) {
	tl := NewTodoList()

	// Normal add
	_, err := tl.Add("Buy milk")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Empty title should error
	_, err = tl.Add("")
	if err == nil {
		t.Error("Empty title should return error")
	}

	// Spaces only - also error
	_, err = tl.Add("   ")
	if err == nil {
		t.Error("Whitespace title should error")
	}

	// Check ID increment
	tasks := tl.GetAll()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != 1 {
		t.Errorf("First task ID should be 1, got %d", tasks[0].ID)
	}
}

// Test retrieving tasks
func TestGetAllTasks(t *testing.T) {
	tl := NewTodoList()

	// Empty list should return empty slice, not nil
	tasks := tl.GetAll()
	if tasks == nil {
		t.Error("GetAll() returned nil, should be empty slice")
	}

	// Add some
	_, err := tl.Add("Task 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = tl.Add("Task 2")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	tasks = tl.GetAll()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

// Test filtering by status
func TestFilterByStatus(t *testing.T) {
	tl := NewTodoList()
	_, err := tl.Add("Task 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = tl.Add("Task 2")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	err = tl.Complete(1)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	pending := tl.GetByStatus(false)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending, got %d", len(pending))
	}

	done := tl.GetByStatus(true)
	if len(done) != 1 {
		t.Errorf("Expected 1 done, got %d", len(done))
	}
}

// Test completion
func TestCompleteTask(t *testing.T) {
	tl := NewTodoList()
	_, err := tl.Add("Test task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Complete existing
	err = tl.Complete(1)
	if err != nil {
		t.Errorf("Complete failed: %v", err)
	}

	task := tl.GetAll()[0]
	if !task.Done {
		t.Error("Task should be marked done")
	}

	// Complete same task again - should just stay done
	err = tl.Complete(1)
	if err != nil {
		t.Errorf("Completing already done task failed: %v", err)
	}

	// Complete non-existent
	err = tl.Complete(999)
	if err == nil {
		t.Error("Completing non-existent task should error")
	}
}

// Test deletion
func TestDeleteTask(t *testing.T) {
	tl := NewTodoList()
	_, err := tl.Add("Task 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = tl.Add("Task 2")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Delete middle
	err = tl.Remove(1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	tasks := tl.GetAll()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("Remaining task should be ID 2, got %d", tasks[0].ID)
	}

	// Delete non-existent
	err = tl.Remove(999)
	if err == nil {
		t.Error("Deleting non-existent should error")
	}
}

// Test editing
func TestEditTask(t *testing.T) {
	tl := NewTodoList()
	_, err := tl.Add("Original title")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Normal edit
	err = tl.Update(1, "New title")
	if err != nil {
		t.Fatalf("Edit failed: %v", err)
	}

	task := tl.GetAll()[0]
	if task.Title != "New title" {
		t.Errorf("Title not updated, got '%s'", task.Title)
	}

	// Empty title
	err = tl.Update(1, "")
	if err == nil {
		t.Error("Empty title should error")
	}

	// Edit non-existent
	err = tl.Update(999, "Whatever")
	if err == nil {
		t.Error("Edit non-existent should error")
	}
}

// Test timestamps are set properly
func TestTimestamps(t *testing.T) {
	tl := NewTodoList()

	// Give a tiny delay to ensure time difference
	time.Sleep(10 * time.Millisecond)

	_, err := tl.Add("Test")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	task := tl.GetAll()[0]

	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}

	// Check that UpdatedAt starts same as CreatedAt
	if !task.UpdatedAt.Equal(task.CreatedAt) {
		t.Error("UpdatedAt should equal CreatedAt initially")
	}

	// Complete and see if UpdatedAt changes
	oldTime := task.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	err = tl.Complete(1)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	task = tl.GetAll()[0]
	if task.UpdatedAt.Equal(oldTime) {
		t.Error("UpdatedAt should change after completion")
	}
}

// Quick edge case - many tasks
func TestManyTasks(t *testing.T) {
	tl := NewTodoList()

	// Add 100 tasks
	for i := 0; i < 100; i++ {
		_, err := tl.Add("Task")
		if err != nil {
			t.Fatalf("Failed at %d: %v", i, err)
		}
	}

	tasks := tl.GetAll()
	if len(tasks) != 100 {
		t.Errorf("Expected 100 tasks, got %d", len(tasks))
	}

	if tasks[99].ID != 100 {
		t.Errorf("Last task ID should be 100, got %d", tasks[99].ID)
	}
}
