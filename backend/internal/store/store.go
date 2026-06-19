// Package store defines the persistence interface used by the HTTP handlers
// and a PostgreSQL implementation of it. Depending on an interface keeps the
// handlers testable with an in-memory fake.
package store

import (
	"context"
	"errors"

	"github.com/hireft/task-manager/internal/models"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrEmailTaken is returned when signing up with an already-registered email.
var ErrEmailTaken = errors.New("email already registered")

// Store is the persistence contract for the application.
type Store interface {
	// Users
	CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)

	// Tasks
	CreateTask(ctx context.Context, userID string, req models.CreateTaskRequest) (models.Task, error)
	GetTask(ctx context.Context, id string) (models.Task, error)
	ListTasks(ctx context.Context, userID string, includeAll bool, f models.TaskFilter) ([]models.Task, int, error)
	UpdateTask(ctx context.Context, id string, req models.UpdateTaskRequest) (models.Task, error)
	DeleteTask(ctx context.Context, id string) error

	// Activity log
	AddActivity(ctx context.Context, taskID, userID, action, detail string) error
	ListActivity(ctx context.Context, taskID string) ([]models.Activity, error)
}
