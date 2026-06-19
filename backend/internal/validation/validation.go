// Package validation contains input-validation rules for write endpoints.
// Each function returns a map of field -> message so the HTTP layer can emit
// a consistent error response.
package validation

import (
	"net/mail"
	"strings"
	"time"

	"github.com/hireft/task-manager/internal/models"
)

const (
	maxTitleLen       = 200
	maxDescriptionLen = 5000
	minPasswordLen    = 8
	maxPasswordLen    = 128
)

var validStatuses = map[string]bool{
	models.StatusTodo:       true,
	models.StatusInProgress: true,
	models.StatusDone:       true,
}

var validPriorities = map[string]bool{
	models.PriorityLow:    true,
	models.PriorityMedium: true,
	models.PriorityHigh:   true,
}

// ValidateSignup checks credentials supplied to signup/login-style endpoints.
func ValidateSignup(req models.SignupRequest) map[string]string {
	errs := map[string]string{}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = "email is required"
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "email is not a valid address"
	}
	if len(req.Password) < minPasswordLen {
		errs["password"] = "password must be at least 8 characters"
	} else if len(req.Password) > maxPasswordLen {
		errs["password"] = "password must be at most 128 characters"
	}
	return errs
}

// ValidateCreateTask validates the body of POST /tasks.
func ValidateCreateTask(req models.CreateTaskRequest) map[string]string {
	errs := map[string]string{}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		errs["title"] = "title is required"
	} else if len(title) > maxTitleLen {
		errs["title"] = "title must be at most 200 characters"
	}

	if len(req.Description) > maxDescriptionLen {
		errs["description"] = "description must be at most 5000 characters"
	}

	// status and priority are optional on create (defaults applied later),
	// but if present they must be valid.
	if req.Status != "" && !validStatuses[req.Status] {
		errs["status"] = "status must be one of: todo, in_progress, done"
	}
	if req.Priority != "" && !validPriorities[req.Priority] {
		errs["priority"] = "priority must be one of: low, medium, high"
	}
	if req.DueDate != nil && req.DueDate.IsZero() {
		errs["dueDate"] = "dueDate must be a valid RFC3339 timestamp"
	}
	return errs
}

// ValidateUpdateTask validates the body of PATCH /tasks/:id. Only provided
// fields are checked. It returns an "_" error if the request is entirely empty.
func ValidateUpdateTask(req models.UpdateTaskRequest) map[string]string {
	errs := map[string]string{}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			errs["title"] = "title cannot be empty"
		} else if len(title) > maxTitleLen {
			errs["title"] = "title must be at most 200 characters"
		}
	}
	if req.Description != nil && len(*req.Description) > maxDescriptionLen {
		errs["description"] = "description must be at most 5000 characters"
	}
	if req.Status != nil && !validStatuses[*req.Status] {
		errs["status"] = "status must be one of: todo, in_progress, done"
	}
	if req.Priority != nil && !validPriorities[*req.Priority] {
		errs["priority"] = "priority must be one of: low, medium, high"
	}

	hasChange := req.Title != nil || req.Description != nil || req.Status != nil ||
		req.Priority != nil || req.DueDate != nil || req.ClearDue
	if len(errs) == 0 && !hasChange {
		errs["_"] = "no fields provided to update"
	}
	return errs
}

// NormalizeFilter parses and bounds the list query parameters, applying
// defaults for sorting and pagination.
func NormalizeFilter(status, search, sortBy, sortDir string, page, pageSize int) models.TaskFilter {
	f := models.TaskFilter{
		Search: strings.TrimSpace(search),
	}

	if validStatuses[status] {
		f.Status = status
	}

	switch sortBy {
	case "due_date", "priority", "created_at":
		f.SortBy = sortBy
	default:
		f.SortBy = "created_at"
	}

	if strings.EqualFold(sortDir, "asc") {
		f.SortDir = "asc"
	} else {
		f.SortDir = "desc"
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	f.Page = page
	f.PageSize = pageSize
	return f
}

// DueDateValid is a small helper kept for clarity/testing.
func DueDateValid(t *time.Time) bool {
	return t == nil || !t.IsZero()
}
