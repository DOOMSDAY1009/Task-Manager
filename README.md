# Task Manager

A full-stack task management application: a **Go** REST API backed by **PostgreSQL**, and a **Next.js** (App Router, TypeScript) frontend with JWT authentication.

> Built for the Full-Stack Developer Assessment. Covers all five required tasks plus several bonus features (dark mode, optimistic UI, role-based access, activity log, Docker one-command setup, and a CI pipeline).

## Live demo

- **Frontend:** https://task-manager-web-h2d7.onrender.com
- **Backend API:** https://task-manager-api-fwp1.onrender.com (e.g. [`/healthz`](https://task-manager-api-fwp1.onrender.com/healthz))

> Hosted on Render's free tier — services sleep after ~15 min of inactivity, so the **first request may take 30–50s** to wake the instance. Just give it a moment and reload.

---

## Features

**Core**
- REST API: `POST/GET/PATCH/DELETE /tasks` with create, list, fetch-one, update, delete
- Filtering by status + offset pagination
- Search by title and sort by due date / priority / created date — all composable with filters
- Input validation on every write endpoint with a consistent error envelope and correct HTTP status codes
- JWT auth (signup + login), bcrypt-hashed passwords, all task routes protected
- Per-user data isolation: users can only see and modify their own tasks
- Auth persisted on the frontend across refreshes (token in `localStorage`, rehydrated via `/auth/me`)
- Loading / empty / error states; responsive layout (mobile + desktop)

**Bonus**
- 🌙 **Dark mode** with a persisted preference (no flash on load)
- ⚡ **Optimistic UI** for completing and deleting tasks, with rollback on failure
- 👑 **Role-based access**: an `admin` role can list all users' tasks via `?scope=all`
- 📝 **Activity log** per task (`GET /tasks/:id/activity`)
- 🐳 **Dockerized**: `docker compose up` brings up db + backend + frontend
- 🤖 **CI**: GitHub Actions runs backend and frontend tests/builds on push

---

## Tech stack

| Layer    | Choice                                          |
| -------- | ----------------------------------------------- |
| Frontend | Next.js 15 (App Router), React 19, TypeScript   |
| Backend  | Go 1.25, standard-library `net/http` router     |
| Database | PostgreSQL 16 (`pgx` driver)                    |
| Auth     | JWT (HS256, `golang-jwt`) + bcrypt              |

---

## Quick start (Docker — recommended)

The fastest way to run everything:

```bash
cp .env.example .env          # optionally edit secrets
docker compose up --build
```

- Frontend → http://localhost:3000
- Backend  → http://localhost:8080
- Postgres → localhost:5432

The backend runs migrations automatically on startup, so the schema is created for you.

---

## Manual setup (without Docker)

### Prerequisites
- Go 1.25+, Node 20+, and a running PostgreSQL 14+

### 1. Database
Create a database and user (or use an existing Postgres):

```sql
CREATE USER taskuser WITH PASSWORD 'taskpass';
CREATE DATABASE taskdb OWNER taskuser;
```

### 2. Backend

```bash
cd backend
cp .env.example .env          # set DATABASE_URL and a JWT_SECRET (>= 16 chars)
go mod download
go run ./cmd/server           # migrations run automatically; listens on :8080
```

### 3. Frontend

```bash
cd frontend
cp .env.example .env.local    # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev                   # http://localhost:3000
```

---

## Environment variables

### Backend (`backend/.env.example`)
| Variable         | Required | Default                  | Description                              |
| ---------------- | -------- | ------------------------ | ---------------------------------------- |
| `DATABASE_URL`   | ✅       | —                        | PostgreSQL connection string             |
| `JWT_SECRET`     | ✅       | —                        | Signing secret (min 16 chars)            |
| `JWT_EXPIRY`     |          | `24h`                    | Token lifetime (Go duration)             |
| `PORT`           |          | `8080`                   | API listen port                          |
| `CORS_ORIGIN`    |          | `http://localhost:3000`  | Allowed frontend origin                  |
| `BCRYPT_COST`    |          | `12`                     | bcrypt cost factor                       |
| `RUN_MIGRATIONS` |          | `true`                   | Apply migrations on startup              |

### Frontend (`frontend/.env.example`)
| Variable              | Default                  | Description                  |
| --------------------- | ------------------------ | ---------------------------- |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080`  | Base URL of the backend API  |

---

## API reference

All `/tasks` routes require an `Authorization: Bearer <token>` header.

| Method | Path                      | Description                                   |
| ------ | ------------------------- | --------------------------------------------- |
| POST   | `/auth/signup`            | Create an account, returns `{ token, user }`  |
| POST   | `/auth/login`             | Log in, returns `{ token, user }`             |
| GET    | `/auth/me`                | Current user (used to rehydrate the session)  |
| POST   | `/tasks`                  | Create a task                                 |
| GET    | `/tasks`                  | List tasks (filter/search/sort/paginate)      |
| GET    | `/tasks/:id`              | Fetch one task                                |
| PATCH  | `/tasks/:id`              | Partial update                                |
| DELETE | `/tasks/:id`              | Delete a task                                 |
| GET    | `/tasks/:id/activity`     | Task change history (bonus)                   |
| GET    | `/healthz`                | Health check                                  |

### List query parameters
`GET /tasks?status=todo&search=report&sortBy=due_date&sortDir=asc&page=1&pageSize=10`

- `status`: `todo` \| `in_progress` \| `done`
- `search`: case-insensitive title match
- `sortBy`: `created_at` (default) \| `due_date` \| `priority`
- `sortDir`: `desc` (default) \| `asc`
- `page`: 1-based (default 1) · `pageSize`: default 10, max 100
- `scope=all`: **admins only** — include every user's tasks

### Error format
Errors use a consistent envelope. Validation failures return `422` with field details:

```json
{ "error": { "message": "validation failed", "fields": { "title": "title is required" } } }
```

### Example

```bash
# Sign up
curl -X POST localhost:8080/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"password123"}'

# Create a task (use the token from the response above)
curl -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Ship the assessment","priority":"high","status":"todo"}'
```

---

## Tests

**Backend** (validation rules, JWT/bcrypt, and full HTTP integration tests against an in-memory store — no database required):

```bash
cd backend && go test ./...
```

**Frontend** (date/overdue helpers via Vitest):

```bash
cd frontend && npm test
```

There are well over the required 3 meaningful tests, including ownership-enforcement and combined search+filter+sort coverage.

---

## Promoting a user to admin (bonus role-based access)

Sign-ups always create a regular `user`. To grant the admin role (which can view all tasks via `?scope=all`), update the row directly:

```sql
UPDATE users SET role = 'admin' WHERE email = 'me@example.com';
```

The new role takes effect on the user's next login (a new token is issued).

---

## Deployment

Two supported paths. Both deploy the backend + Postgres on Render and the frontend on either Render or Vercel.

### Option A — Render Blueprint (backend + db + frontend, one repo)
1. Push this repo to GitHub.
2. In Render: **New → Blueprint** and select the repo. `render.yaml` provisions the database, the Go API, and the Next.js app.
3. After the first deploy, set the two cross-service URLs (the only `sync: false` vars):
   - `task-manager-api` → `CORS_ORIGIN` = the deployed frontend URL
   - `task-manager-web` → `NEXT_PUBLIC_API_URL` = the deployed API URL
   Then redeploy the frontend so the API URL is baked into the build.

### Option B — Vercel (frontend) + Render (backend + db)
1. Deploy the backend with the Render Blueprint (or any host that runs the `backend/Dockerfile`); note the API URL.
2. On Vercel: import the repo, set **Root Directory = `frontend`**, and add env var `NEXT_PUBLIC_API_URL` = the API URL. `frontend/vercel.json` pins the framework and commands.
3. Set the backend's `CORS_ORIGIN` to the Vercel URL.

`JWT_SECRET` is generated automatically by the Render Blueprint; set it manually (16+ chars) on any other host.

## Project structure

```
task-manager/
├── backend/                  # Go REST API
│   ├── cmd/server/           # main entrypoint (graceful shutdown)
│   └── internal/
│       ├── auth/             # bcrypt + JWT
│       ├── config/           # env loading
│       ├── database/         # pool + embedded migrations
│       ├── handlers/         # HTTP handlers + router
│       ├── middleware/       # auth, CORS, logging, recovery
│       ├── models/           # domain + request/response types
│       ├── store/            # Store interface + Postgres & in-memory impls
│       └── validation/       # input validation
├── frontend/                 # Next.js App Router app
│   ├── app/                  # routes (login, signup, tasks)
│   ├── components/           # UI components
│   ├── context/              # auth + theme providers
│   └── lib/                  # api client, types, helpers
├── .github/workflows/ci.yml  # CI pipeline
└── docker-compose.yml        # one-command local setup
```

---

## Assumptions & trade-offs

- **JWT in `localStorage`.** Simple and survives refreshes, which the brief asks for. The trade-off is XSS exposure; a hardened setup would use httpOnly refresh-token cookies. Documented here as a conscious choice for scope.
- **Stateless JWTs, no server-side revocation.** Logout clears the client token; tokens remain valid until expiry (default 24h). A denylist/refresh-token rotation would be the next step.
- **`Store` interface over the database.** Handlers depend on an interface, so the full HTTP test suite runs against an in-memory store with no Postgres required — fast, hermetic CI. The Postgres implementation is exercised via `docker compose`.
- **404 instead of 403** when accessing another user's task, to avoid leaking which task IDs exist.
- **Admin promotion via SQL** rather than an admin-management UI/endpoint, to keep auth surface small for the assessment.
- **Migrations run on startup** (embedded in the binary) for one-command setup; for larger teams a dedicated migration step in the deploy pipeline is preferable.
- **Hand-written CSS with variables** instead of a UI framework, to keep the bundle small and the theming (dark mode) explicit and dependency-free.
