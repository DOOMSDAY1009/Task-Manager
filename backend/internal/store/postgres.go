package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hireft/task-manager/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres is a Store backed by a pgx connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres wraps an existing pool.
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (s *Postgres) CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, role, created_at`,
		strings.ToLower(email), passwordHash, role,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *Postgres) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`,
		strings.ToLower(email),
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (s *Postgres) GetUserByID(ctx context.Context, id string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (s *Postgres) CreateTask(ctx context.Context, userID string, req models.CreateTaskRequest) (models.Task, error) {
	status := req.Status
	if status == "" {
		status = models.StatusTodo
	}
	priority := req.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}

	var t models.Task
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, title, description, status, priority, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at`,
		userID, strings.TrimSpace(req.Title), req.Description, status, priority, req.DueDate,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (s *Postgres) GetTask(ctx context.Context, id string) (models.Task, error) {
	var t models.Task
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Task{}, ErrNotFound
	}
	return t, err
}

func (s *Postgres) ListTasks(ctx context.Context, userID string, includeAll bool, f models.TaskFilter) ([]models.Task, int, error) {
	var where []string
	var args []interface{}
	i := 1

	if !includeAll {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, userID)
		i++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", i))
		args = append(args, "%"+f.Search+"%")
		i++
	}

	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}

	// Total count for pagination metadata.
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks "+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := s.orderClause(f.SortBy, f.SortDir)
	limit := f.PageSize
	offset := (f.Page - 1) * f.PageSize

	query := fmt.Sprintf(
		`SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
		 FROM tasks %s %s LIMIT $%d OFFSET $%d`,
		clause, orderBy, i, i+1,
	)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

// orderClause builds a safe ORDER BY from a whitelisted column and direction.
func (s *Postgres) orderClause(sortBy, dir string) string {
	col := "created_at"
	switch sortBy {
	case "due_date":
		// NULLS LAST keeps undated tasks at the end regardless of direction.
		col = "due_date"
		return fmt.Sprintf("ORDER BY due_date %s NULLS LAST, created_at DESC", sqlDir(dir))
	case "priority":
		// Order by semantic weight, not alphabetically.
		return fmt.Sprintf(
			"ORDER BY CASE priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 ELSE 0 END %s, created_at DESC",
			sqlDir(dir),
		)
	case "created_at":
		col = "created_at"
	}
	return fmt.Sprintf("ORDER BY %s %s", col, sqlDir(dir))
}

func sqlDir(dir string) string {
	if strings.EqualFold(dir, "asc") {
		return "ASC"
	}
	return "DESC"
}

func (s *Postgres) UpdateTask(ctx context.Context, id string, req models.UpdateTaskRequest) (models.Task, error) {
	var sets []string
	var args []interface{}
	i := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", i))
		args = append(args, strings.TrimSpace(*req.Title))
		i++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *req.Description)
		i++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *req.Status)
		i++
	}
	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", i))
		args = append(args, *req.Priority)
		i++
	}
	if req.ClearDue {
		sets = append(sets, "due_date = NULL")
	} else if req.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date = $%d", i))
		args = append(args, *req.DueDate)
		i++
	}

	if len(sets) == 0 {
		// Nothing to change; return the current row.
		return s.GetTask(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE tasks SET %s WHERE id = $%d
		 RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at`,
		strings.Join(sets, ", "), i,
	)

	var t models.Task
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Task{}, ErrNotFound
	}
	return t, err
}

func (s *Postgres) DeleteTask(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Postgres) AddActivity(ctx context.Context, taskID, userID, action, detail string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO activity_log (task_id, user_id, action, detail) VALUES ($1, $2, $3, $4)`,
		taskID, userID, action, detail,
	)
	return err
}

func (s *Postgres) ListActivity(ctx context.Context, taskID string) ([]models.Activity, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, task_id, user_id, action, detail, created_at
		 FROM activity_log WHERE task_id = $1 ORDER BY created_at DESC`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Activity{}
	for rows.Next() {
		var a models.Activity
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Action, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}
