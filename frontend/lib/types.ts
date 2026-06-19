// Shared domain types mirroring the backend JSON shapes.

export type TaskStatus = "todo" | "in_progress" | "done";
export type TaskPriority = "low" | "medium" | "high";
export type Role = "user" | "admin";

export interface User {
  id: string;
  email: string;
  role: Role;
  createdAt: string;
}

export interface Task {
  id: string;
  userId: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueDate: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedTasks {
  tasks: Task[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface CreateTaskInput {
  title: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  dueDate?: string | null;
}

export interface UpdateTaskInput {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  dueDate?: string | null;
  clearDueDate?: boolean;
}

export interface ListParams {
  status?: TaskStatus | "";
  search?: string;
  sortBy?: "due_date" | "priority" | "created_at";
  sortDir?: "asc" | "desc";
  page?: number;
  pageSize?: number;
}

// Error envelope returned by the API.
export interface ApiError {
  message: string;
  fields?: Record<string, string>;
}
