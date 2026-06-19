// Package models defines the core domain types and the request/response
// shapes used by the HTTP layer.
package models

import "time"

// Role represents a user's authorization level.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Task status values.
const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// Task priority values.
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// User is an application account.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Task is a unit of work owned by a user.
type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Activity records a change made to a task (bonus: activity log).
type Activity struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	UserID    string    `json:"userId"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

// --- Request payloads ---

// SignupRequest is the body for POST /auth/signup.
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateTaskRequest is the body for POST /tasks.
type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
}

// UpdateTaskRequest is the body for PATCH /tasks/:id. Pointer fields allow
// distinguishing "not provided" from "set to zero value" for partial updates.
type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Priority    *string    `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	ClearDue    bool       `json:"clearDueDate"`
}

// TaskFilter captures the query parameters for GET /tasks.
type TaskFilter struct {
	Status   string // optional exact-match status filter
	Search   string // optional case-insensitive title search
	SortBy   string // due_date | priority | created_at
	SortDir  string // asc | desc
	Page     int    // 1-based
	PageSize int
}

// --- Response payloads ---

// AuthResponse is returned by signup and login.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// PaginatedTasks is the envelope for GET /tasks.
type PaginatedTasks struct {
	Tasks      []Task `json:"tasks"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}
