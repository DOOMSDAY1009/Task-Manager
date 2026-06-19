package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hireft/task-manager/internal/middleware"
	"github.com/hireft/task-manager/internal/models"
	"github.com/hireft/task-manager/internal/store"
	"github.com/hireft/task-manager/internal/validation"
)

// currentUser pulls the authenticated identity off the request context.
func (h *Handler) currentUser(r *http.Request) (middleware.AuthUser, error) {
	u, ok := middleware.FromContext(r.Context())
	if !ok {
		return middleware.AuthUser{}, errors.New("no user in context")
	}
	return u, nil
}

// CreateTask handles POST /tasks.
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req models.CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if errs := validation.ValidateCreateTask(req); len(errs) > 0 {
		writeValidationError(w, errs)
		return
	}

	task, err := h.Store.CreateTask(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create task")
		return
	}
	_ = h.Store.AddActivity(r.Context(), task.ID, user.ID, "created", "task created")
	writeJSON(w, http.StatusCreated, task)
}

// ListTasks handles GET /tasks with filtering, search, sort, and pagination.
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	filter := validation.NormalizeFilter(
		q.Get("status"), q.Get("search"), q.Get("sortBy"), q.Get("sortDir"), page, pageSize,
	)

	// Admins may opt into viewing all users' tasks with ?scope=all.
	includeAll := user.Role == models.RoleAdmin && q.Get("scope") == "all"

	tasks, total, err := h.Store.ListTasks(r.Context(), user.ID, includeAll, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list tasks")
		return
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize
	writeJSON(w, http.StatusOK, models.PaginatedTasks{
		Tasks:      tasks,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	})
}

// GetTask handles GET /tasks/{id}.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	user, task, ok := h.loadOwnedTask(w, r)
	if !ok {
		return
	}
	_ = user
	writeJSON(w, http.StatusOK, task)
}

// UpdateTask handles PATCH /tasks/{id}.
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	user, _, ok := h.loadOwnedTask(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")

	var req models.UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if errs := validation.ValidateUpdateTask(req); len(errs) > 0 {
		writeValidationError(w, errs)
		return
	}

	updated, err := h.Store.UpdateTask(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update task")
		return
	}
	_ = h.Store.AddActivity(r.Context(), id, user.ID, "updated", "task updated")
	writeJSON(w, http.StatusOK, updated)
}

// DeleteTask handles DELETE /tasks/{id}.
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	user, _, ok := h.loadOwnedTask(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")

	if err := h.Store.DeleteTask(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete task")
		return
	}
	_ = user
	w.WriteHeader(http.StatusNoContent)
}

// GetActivity handles GET /tasks/{id}/activity (bonus: activity log).
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	_, _, ok := h.loadOwnedTask(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	activity, err := h.Store.ListActivity(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load activity")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"activity": activity})
}

// loadOwnedTask fetches the task referenced by {id} and enforces that the
// current user owns it (admins bypass the ownership check). It writes the
// appropriate error response and returns ok=false on any failure.
func (h *Handler) loadOwnedTask(w http.ResponseWriter, r *http.Request) (middleware.AuthUser, models.Task, bool) {
	user, err := h.currentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return middleware.AuthUser{}, models.Task{}, false
	}

	id := r.PathValue("id")
	task, err := h.Store.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return middleware.AuthUser{}, models.Task{}, false
		}
		writeError(w, http.StatusInternalServerError, "could not load task")
		return middleware.AuthUser{}, models.Task{}, false
	}

	if task.UserID != user.ID && user.Role != models.RoleAdmin {
		// Return 404 rather than 403 so we don't reveal that the task exists.
		writeError(w, http.StatusNotFound, "task not found")
		return middleware.AuthUser{}, models.Task{}, false
	}
	return user, task, true
}
