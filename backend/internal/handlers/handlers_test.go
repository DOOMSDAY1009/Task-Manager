package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hireft/task-manager/internal/auth"
	"github.com/hireft/task-manager/internal/handlers"
	"github.com/hireft/task-manager/internal/models"
	"github.com/hireft/task-manager/internal/store"
)

// newTestServer wires the real router against an in-memory store.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	st := store.NewMemory()
	mgr := auth.NewManager("integration-test-secret-key", time.Hour)
	h := handlers.New(st, mgr, 4) // low bcrypt cost for speed
	return httptest.NewServer(handlers.Routes(h, mgr, "*"))
}

// do is a small helper for issuing JSON requests with an optional bearer token.
func do(t *testing.T, srv *httptest.Server, method, path, token string, body interface{}) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func signup(t *testing.T, srv *httptest.Server, email, password string) string {
	t.Helper()
	resp := do(t, srv, http.MethodPost, "/auth/signup", "", map[string]string{
		"email": email, "password": password,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("signup status = %d, want 201", resp.StatusCode)
	}
	var out models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Token == "" {
		t.Fatal("signup returned empty token")
	}
	return out.Token
}

func TestAuthFlowAndProtectedRoutes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Unauthenticated access is rejected.
	resp := do(t, srv, http.MethodGet, "/tasks", "", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET /tasks = %d, want 401", resp.StatusCode)
	}

	token := signup(t, srv, "alice@example.com", "password123")

	// Duplicate signup returns 409.
	dup := do(t, srv, http.MethodPost, "/auth/signup", "", map[string]string{
		"email": "alice@example.com", "password": "password123",
	})
	dup.Body.Close()
	if dup.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate signup = %d, want 409", dup.StatusCode)
	}

	// Login with correct and incorrect credentials.
	ok := do(t, srv, http.MethodPost, "/auth/login", "", map[string]string{
		"email": "alice@example.com", "password": "password123",
	})
	ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("login = %d, want 200", ok.StatusCode)
	}
	bad := do(t, srv, http.MethodPost, "/auth/login", "", map[string]string{
		"email": "alice@example.com", "password": "wrong",
	})
	bad.Body.Close()
	if bad.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad login = %d, want 401", bad.StatusCode)
	}

	_ = token
}

func TestTaskCRUDAndValidation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	token := signup(t, srv, "bob@example.com", "password123")

	// Validation failure: empty title -> 422.
	bad := do(t, srv, http.MethodPost, "/tasks", token, map[string]string{"title": ""})
	bad.Body.Close()
	if bad.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty title = %d, want 422", bad.StatusCode)
	}

	// Create a task.
	resp := do(t, srv, http.MethodPost, "/tasks", token, map[string]interface{}{
		"title": "Write tests", "priority": "high", "status": "todo",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task = %d, want 201", resp.StatusCode)
	}
	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()
	if task.ID == "" || task.Title != "Write tests" || task.Priority != "high" {
		t.Fatalf("unexpected task: %+v", task)
	}

	// Update via PATCH.
	upd := do(t, srv, http.MethodPatch, "/tasks/"+task.ID, token, map[string]string{"status": "done"})
	var updated models.Task
	json.NewDecoder(upd.Body).Decode(&updated)
	upd.Body.Close()
	if upd.StatusCode != http.StatusOK || updated.Status != "done" {
		t.Fatalf("patch failed: status=%d task=%+v", upd.StatusCode, updated)
	}

	// List should contain exactly one task.
	list := do(t, srv, http.MethodGet, "/tasks", token, nil)
	var page models.PaginatedTasks
	json.NewDecoder(list.Body).Decode(&page)
	list.Body.Close()
	if page.Total != 1 || len(page.Tasks) != 1 {
		t.Fatalf("expected 1 task, got total=%d len=%d", page.Total, len(page.Tasks))
	}

	// Delete returns 204, then GET returns 404.
	del := do(t, srv, http.MethodDelete, "/tasks/"+task.ID, token, nil)
	del.Body.Close()
	if del.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d, want 204", del.StatusCode)
	}
	gone := do(t, srv, http.MethodGet, "/tasks/"+task.ID, token, nil)
	gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted = %d, want 404", gone.StatusCode)
	}
}

func TestUsersCannotAccessOthersTasks(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	aliceToken := signup(t, srv, "alice2@example.com", "password123")
	bobToken := signup(t, srv, "bob2@example.com", "password123")

	// Alice creates a task.
	resp := do(t, srv, http.MethodPost, "/tasks", aliceToken, map[string]string{"title": "Alice secret"})
	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	// Bob cannot read it (404, not 403, to avoid leaking existence).
	get := do(t, srv, http.MethodGet, "/tasks/"+task.ID, bobToken, nil)
	get.Body.Close()
	if get.StatusCode != http.StatusNotFound {
		t.Fatalf("bob reading alice's task = %d, want 404", get.StatusCode)
	}

	// Bob cannot update or delete it either.
	upd := do(t, srv, http.MethodPatch, "/tasks/"+task.ID, bobToken, map[string]string{"status": "done"})
	upd.Body.Close()
	if upd.StatusCode != http.StatusNotFound {
		t.Fatalf("bob updating alice's task = %d, want 404", upd.StatusCode)
	}

	// Bob's own list is empty.
	list := do(t, srv, http.MethodGet, "/tasks", bobToken, nil)
	var page models.PaginatedTasks
	json.NewDecoder(list.Body).Decode(&page)
	list.Body.Close()
	if page.Total != 0 {
		t.Fatalf("bob should see 0 tasks, got %d", page.Total)
	}
}

func TestSearchFilterSortWorkTogether(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	token := signup(t, srv, "carol@example.com", "password123")

	seed := []map[string]interface{}{
		{"title": "Buy groceries", "status": "todo", "priority": "low"},
		{"title": "Buy a car", "status": "done", "priority": "high"},
		{"title": "Read a book", "status": "todo", "priority": "medium"},
	}
	for _, s := range seed {
		r := do(t, srv, http.MethodPost, "/tasks", token, s)
		r.Body.Close()
	}

	// search=buy + status=todo should match only "Buy groceries".
	list := do(t, srv, http.MethodGet, "/tasks?search=buy&status=todo&sortBy=priority&sortDir=desc", token, nil)
	var page models.PaginatedTasks
	json.NewDecoder(list.Body).Decode(&page)
	list.Body.Close()
	if page.Total != 1 || page.Tasks[0].Title != "Buy groceries" {
		t.Fatalf("combined search+filter failed: %+v", page)
	}
}
