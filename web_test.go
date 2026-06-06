package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestHandleIndexFilters(t *testing.T) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	todo := NewTodoList()
	_, _ = todo.Add("First task")
	_, _ = todo.Add("Second task")
	_ = todo.Complete(1)
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodGet, "/?filter=all", nil)
	rec := httptest.NewRecorder()
	ws.handleIndex(tmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for index all, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("First task")) || !bytes.Contains(rec.Body.Bytes(), []byte("Second task")) {
		t.Error("Expected both tasks to render for all filter")
	}

	req = httptest.NewRequest(http.MethodGet, "/?filter=pending", nil)
	rec = httptest.NewRecorder()
	ws.handleIndex(tmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for pending filter, got %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("First task")) {
		t.Error("Expected completed task to be hidden in pending filter")
	}
}

func TestHandleTaskDetail(t *testing.T) {
	detailTmpl := template.Must(template.ParseFiles("detail.html"))
	todo := NewTodoList()
	_, _ = todo.Add("Detail task")
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodGet, "/task/1", nil)
	rec := httptest.NewRecorder()
	ws.handleTaskDetail(detailTmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for task detail, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Detail task")) {
		t.Error("Expected task detail page to include task title")
	}

	req = httptest.NewRequest(http.MethodGet, "/task/999", nil)
	rec = httptest.NewRecorder()
	ws.handleTaskDetail(detailTmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for missing task detail, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Task Not Found")) {
		t.Error("Expected missing task detail page to show Task Not Found")
	}
}

func TestHandleAPIGetPostExport(t *testing.T) {
	todo := NewTodoList()
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	ws.handleAPI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for API GET, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "[]") {
		t.Error("Expected empty task list JSON for API GET")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"API task"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	ws.handleAPI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for API POST, got %d", rec.Code)
	}
	var apiResponse map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &apiResponse); err != nil {
		t.Fatalf("Expected valid JSON response from API POST, got %v", err)
	}
	if apiResponse["id"] != 1 {
		t.Fatalf("Expected API POST response id 1, got %d", apiResponse["id"])
	}

	if len(todo.GetAll()) != 1 {
		t.Fatalf("Expected 1 task after API POST, got %d", len(todo.GetAll()))
	}

	req = httptest.NewRequest(http.MethodGet, "/export/json", nil)
	rec = httptest.NewRecorder()
	ws.handleExport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for export, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "todo_export_") {
		t.Error("Expected export response to include file name")
	}
	var exported TodoList
	if err := json.Unmarshal(rec.Body.Bytes(), &exported); err != nil {
		t.Fatalf("Unmarshal export JSON failed: %v", err)
	}
	if len(exported.Tasks) != 1 {
		t.Fatalf("Expected 1 exported task, got %d", len(exported.Tasks))
	}
}

func TestWebHandlerErrorBranches(t *testing.T) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	detailTmpl := template.Must(template.ParseFiles("detail.html"))
	todo := NewTodoList()
	_, _ = todo.Add("Error branch task")
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	rec := httptest.NewRecorder()
	ws.handleAdd(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405 for invalid add method, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/task/bad", nil)
	rec = httptest.NewRecorder()
	ws.handleTaskDetail(detailTmpl)(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for invalid detail id, got %d", rec.Code)
	}

	form := url.Values{}
	form.Set("id", "999")
	req = httptest.NewRequest(http.MethodPost, "/complete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ws.handleComplete(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for complete missing id, got %d", rec.Code)
	}

	form = url.Values{}
	form.Set("id", "1")
	form.Set("title", "")
	req = httptest.NewRequest(http.MethodPost, "/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ws.handleEdit(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for edit empty title, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/delete", nil)
	rec = httptest.NewRecorder()
	ws.handleDelete(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405 for invalid delete method, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/tasks", nil)
	rec = httptest.NewRecorder()
	ws.handleAPI(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405 for invalid API method, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/?filter=completed", nil)
	rec = httptest.NewRecorder()
	ws.handleIndex(tmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 for completed filter, got %d", rec.Code)
	}
}

func TestHandleTaskDetailCompletedTask(t *testing.T) {
	detailTmpl := template.Must(template.ParseFiles("detail.html"))
	todo := NewTodoList()
	_, _ = todo.Add("Completed task")
	_ = todo.Complete(1)
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodGet, "/task/1", nil)
	rec := httptest.NewRecorder()
	ws.handleTaskDetail(detailTmpl)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for completed task detail, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Time to Complete")) {
		t.Error("Expected completed task detail page to include time to complete")
	}
}

func TestHandleCompleteWithReferer(t *testing.T) {
	todo := NewTodoList()
	_, _ = todo.Add("Referer task")
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("id", "1")
	req := httptest.NewRequest(http.MethodPost, "/complete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://localhost/task/1")
	rec := httptest.NewRecorder()

	ws.handleComplete(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Expected 303 redirect for complete with referer, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "http://localhost/task/1" {
		t.Fatalf("Expected redirect to referer, got %q", rec.Header().Get("Location"))
	}
}

func TestHandleEditWithReferer(t *testing.T) {
	todo := NewTodoList()
	_, _ = todo.Add("Referer edit task")
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("id", "1")
	form.Set("title", "Edited via referer")
	req := httptest.NewRequest(http.MethodPost, "/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://localhost/task/1")
	rec := httptest.NewRecorder()

	ws.handleEdit(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Expected 303 redirect for edit with referer, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "http://localhost/task/1" {
		t.Fatalf("Expected redirect to referer, got %q", rec.Header().Get("Location"))
	}
}

func TestHandleAddEmptyTitle(t *testing.T) {
	todo := NewTodoList()
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("title", "")
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ws.handleAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for empty add title, got %d", rec.Code)
	}
}

func TestHandleAPIInvalidJSON(t *testing.T) {
	todo := NewTodoList()
	ws := NewWebServer(todo)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ws.handleAPI(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for invalid JSON API POST, got %d", rec.Code)
	}
}

func TestHandleDeleteInvalidID(t *testing.T) {
	todo := NewTodoList()
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("id", "abc")
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ws.handleDelete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for delete invalid ID, got %d", rec.Code)
	}
}

func TestHandleEditInvalidID(t *testing.T) {
	todo := NewTodoList()
	_, _ = todo.Add("Task")
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("id", "abc")
	form.Set("title", "New title")
	req := httptest.NewRequest(http.MethodPost, "/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ws.handleEdit(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for edit invalid ID, got %d", rec.Code)
	}
}

func TestHandleEditNoRefererRedirectsHome(t *testing.T) {
	todo := NewTodoList()
	_, _ = todo.Add("Task no referer")
	ws := NewWebServer(todo)

	form := url.Values{}
	form.Set("id", "1")
	form.Set("title", "Updated Home")
	req := httptest.NewRequest(http.MethodPost, "/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ws.handleEdit(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Expected 303 for edit no referer, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/" {
		t.Fatalf("Expected redirect to home, got %q", rec.Header().Get("Location"))
	}
}

func TestWebServerStartRegistersRoutes(t *testing.T) {
	originalMux := http.DefaultServeMux
	originalListen := listenAndServe
	defer func() {
		http.DefaultServeMux = originalMux
		listenAndServe = originalListen
	}()

	http.DefaultServeMux = http.NewServeMux()
	listenAndServe = func(addr string, handler http.Handler) error {
		if addr != ":12345" {
			t.Fatalf("Expected port :12345, got %s", addr)
		}
		if handler == nil {
			handler = http.DefaultServeMux
		}
		req := httptest.NewRequest(http.MethodGet, "/?filter=all", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected registered handler response 200, got %d", rec.Code)
		}
		return nil
	}

	todo := NewTodoList()
	ws := NewWebServer(todo)
	if err := ws.Start(":12345"); err != nil {
		t.Fatalf("Start should not return error: %v", err)
	}
}

func TestMainCLIFlow(t *testing.T) {
	oldArgs := os.Args
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldWd, _ := os.Getwd()
	defer func() {
		os.Args = oldArgs
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Chdir(oldWd)
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, err = wIn.WriteString("add task one\nlist all\ndone 1\nedit 1 task one updated\ndel 1\nexit\n")
	if err != nil {
		t.Fatal(err)
	}
	wIn.Close()
	os.Stdin = rIn

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wOut

	os.Args = []string{"cmd"}
	main()
	wOut.Close()

	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Goodbye!") {
		t.Errorf("Expected main CLI to print Goodbye, got %q", string(out))
	}
}

func TestMainWebBranch(t *testing.T) {
	oldArgs := os.Args
	oldWd, _ := os.Getwd()
	oldListen := listenAndServe
	oldMux := http.DefaultServeMux
	defer func() {
		os.Args = oldArgs
		os.Chdir(oldWd)
		listenAndServe = oldListen
		http.DefaultServeMux = oldMux
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("index.html", []byte(`<!DOCTYPE html><html><body>{{range .Tasks}}{{.Title}}{{end}}</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("detail.html", []byte(`<!DOCTYPE html><html><body>{{if .Task}}{{.Task.Title}}{{else}}Task Not Found{{end}}</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}

	http.DefaultServeMux = http.NewServeMux()
	listenAndServe = func(addr string, handler http.Handler) error {
		if addr != ":12346" {
			t.Fatalf("Expected web port :12346, got %s", addr)
		}
		if handler == nil {
			handler = http.DefaultServeMux
		}
		req := httptest.NewRequest(http.MethodGet, "/?filter=all", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 from registered route, got %d", rec.Code)
		}
		return nil
	}

	os.Args = []string{"cmd", "--web", "12346"}
	main()
}
