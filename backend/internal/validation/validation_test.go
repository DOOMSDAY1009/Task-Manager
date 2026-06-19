package validation

import (
	"testing"

	"github.com/hireft/task-manager/internal/models"
)

func ptr[T any](v T) *T { return &v }

func TestValidateSignup(t *testing.T) {
	cases := []struct {
		name      string
		req       models.SignupRequest
		wantField string // "" means expect no errors
	}{
		{"valid", models.SignupRequest{Email: "a@b.com", Password: "password1"}, ""},
		{"missing email", models.SignupRequest{Password: "password1"}, "email"},
		{"bad email", models.SignupRequest{Email: "not-an-email", Password: "password1"}, "email"},
		{"short password", models.SignupRequest{Email: "a@b.com", Password: "short"}, "password"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := ValidateSignup(c.req)
			if c.wantField == "" {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if _, ok := errs[c.wantField]; !ok {
				t.Fatalf("expected error on field %q, got %v", c.wantField, errs)
			}
		})
	}
}

func TestValidateCreateTask(t *testing.T) {
	cases := []struct {
		name      string
		req       models.CreateTaskRequest
		wantField string
	}{
		{"valid minimal", models.CreateTaskRequest{Title: "Buy milk"}, ""},
		{"valid full", models.CreateTaskRequest{Title: "X", Status: "done", Priority: "high"}, ""},
		{"empty title", models.CreateTaskRequest{Title: "   "}, "title"},
		{"bad status", models.CreateTaskRequest{Title: "X", Status: "nope"}, "status"},
		{"bad priority", models.CreateTaskRequest{Title: "X", Priority: "urgent"}, "priority"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := ValidateCreateTask(c.req)
			if c.wantField == "" && len(errs) != 0 {
				t.Fatalf("expected no errors, got %v", errs)
			}
			if c.wantField != "" {
				if _, ok := errs[c.wantField]; !ok {
					t.Fatalf("expected error on %q, got %v", c.wantField, errs)
				}
			}
		})
	}
}

func TestValidateUpdateTaskRejectsEmptyBody(t *testing.T) {
	errs := ValidateUpdateTask(models.UpdateTaskRequest{})
	if _, ok := errs["_"]; !ok {
		t.Fatalf("expected empty-update error, got %v", errs)
	}
}

func TestValidateUpdateTaskRejectsBlankTitle(t *testing.T) {
	errs := ValidateUpdateTask(models.UpdateTaskRequest{Title: ptr("  ")})
	if _, ok := errs["title"]; !ok {
		t.Fatalf("expected title error, got %v", errs)
	}
}

func TestNormalizeFilterDefaultsAndBounds(t *testing.T) {
	f := NormalizeFilter("garbage", "  hello  ", "weird", "ASC", 0, 9999)
	if f.Status != "" {
		t.Errorf("invalid status should be dropped, got %q", f.Status)
	}
	if f.Search != "hello" {
		t.Errorf("search should be trimmed, got %q", f.Search)
	}
	if f.SortBy != "created_at" {
		t.Errorf("invalid sortBy should default to created_at, got %q", f.SortBy)
	}
	if f.SortDir != "asc" {
		t.Errorf("expected asc, got %q", f.SortDir)
	}
	if f.Page != 1 {
		t.Errorf("page 0 should clamp to 1, got %d", f.Page)
	}
	if f.PageSize != 100 {
		t.Errorf("pageSize should clamp to 100, got %d", f.PageSize)
	}
}
