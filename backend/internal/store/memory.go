package store

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hireft/task-manager/internal/models"
)

// Memory is an in-memory Store implementation used for tests and as a simple
// fallback. It is safe for concurrent use.
type Memory struct {
	mu       sync.RWMutex
	users    map[string]models.User
	tasks    map[string]models.Task
	activity []models.Activity
}

func nowUTC() time.Time { return time.Now().UTC() }

// NewMemory builds an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{
		users: map[string]models.User{},
		tasks: map[string]models.Task{},
	}
}

func (m *Memory) CreateUser(_ context.Context, email, passwordHash string, role models.Role) (models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	email = strings.ToLower(email)
	for _, u := range m.users {
		if u.Email == email {
			return models.User{}, ErrEmailTaken
		}
	}
	u := models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    nowUTC(),
	}
	m.users[u.ID] = u
	return u, nil
}

func (m *Memory) GetUserByEmail(_ context.Context, email string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	email = strings.ToLower(email)
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

func (m *Memory) GetUserByID(_ context.Context, id string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return u, nil
}

func (m *Memory) CreateTask(_ context.Context, userID string, req models.CreateTaskRequest) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := req.Status
	if status == "" {
		status = models.StatusTodo
	}
	priority := req.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}
	now := nowUTC()
	t := models.Task{
		ID:          uuid.NewString(),
		UserID:      userID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.tasks[t.ID] = t
	return t, nil
}

func (m *Memory) GetTask(_ context.Context, id string) (models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func (m *Memory) ListTasks(_ context.Context, userID string, includeAll bool, f models.TaskFilter) ([]models.Task, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []models.Task
	for _, t := range m.tasks {
		if !includeAll && t.UserID != userID {
			continue
		}
		if f.Status != "" && t.Status != f.Status {
			continue
		}
		if f.Search != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(f.Search)) {
			continue
		}
		filtered = append(filtered, t)
	}

	sortTasks(filtered, f.SortBy, f.SortDir)
	total := len(filtered)

	start := (f.Page - 1) * f.PageSize
	if start > total {
		start = total
	}
	end := start + f.PageSize
	if end > total {
		end = total
	}
	page := filtered[start:end]
	if page == nil {
		page = []models.Task{}
	}
	return page, total, nil
}

func sortTasks(tasks []models.Task, sortBy, dir string) {
	asc := strings.EqualFold(dir, "asc")
	priorityWeight := map[string]int{models.PriorityLow: 1, models.PriorityMedium: 2, models.PriorityHigh: 3}

	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		var less bool
		switch sortBy {
		case "due_date":
			// nil due dates sort last in both directions.
			switch {
			case a.DueDate == nil && b.DueDate == nil:
				return a.CreatedAt.After(b.CreatedAt)
			case a.DueDate == nil:
				return false
			case b.DueDate == nil:
				return true
			default:
				less = a.DueDate.Before(*b.DueDate)
			}
		case "priority":
			less = priorityWeight[a.Priority] < priorityWeight[b.Priority]
		default: // created_at
			less = a.CreatedAt.Before(b.CreatedAt)
		}
		if asc {
			return less
		}
		return !less
	})
}

func (m *Memory) UpdateTask(_ context.Context, id string, req models.UpdateTaskRequest) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	if req.Priority != nil {
		t.Priority = *req.Priority
	}
	if req.ClearDue {
		t.DueDate = nil
	} else if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
	t.UpdatedAt = nowUTC()
	m.tasks[id] = t
	return t, nil
}

func (m *Memory) DeleteTask(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *Memory) AddActivity(_ context.Context, taskID, userID, action, detail string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activity = append(m.activity, models.Activity{
		ID:        uuid.NewString(),
		TaskID:    taskID,
		UserID:    userID,
		Action:    action,
		Detail:    detail,
		CreatedAt: nowUTC(),
	})
	return nil
}

func (m *Memory) ListActivity(_ context.Context, taskID string) ([]models.Activity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []models.Activity{}
	for _, a := range m.activity {
		if a.TaskID == taskID {
			out = append(out, a)
		}
	}
	return out, nil
}
