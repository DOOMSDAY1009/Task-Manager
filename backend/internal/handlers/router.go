package handlers

import (
	"net/http"

	"github.com/hireft/task-manager/internal/auth"
	"github.com/hireft/task-manager/internal/middleware"
)

// Routes builds the application's HTTP handler with all middleware applied.
func Routes(h *Handler, mgr *auth.Manager, corsOrigin string) http.Handler {
	mux := http.NewServeMux()

	// Health check (unauthenticated).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth endpoints (unauthenticated).
	mux.HandleFunc("POST /auth/signup", h.Signup)
	mux.HandleFunc("POST /auth/login", h.Login)

	// Protected endpoints.
	requireAuth := middleware.RequireAuth(mgr)
	protected := http.NewServeMux()
	protected.HandleFunc("GET /auth/me", h.Me)
	protected.HandleFunc("POST /tasks", h.CreateTask)
	protected.HandleFunc("GET /tasks", h.ListTasks)
	protected.HandleFunc("GET /tasks/{id}", h.GetTask)
	protected.HandleFunc("PATCH /tasks/{id}", h.UpdateTask)
	protected.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)
	protected.HandleFunc("GET /tasks/{id}/activity", h.GetActivity)

	mux.Handle("/", requireAuth(protected))

	return middleware.Chain(mux,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS(corsOrigin),
	)
}
