"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth";
import { api, ApiClientError } from "@/lib/api";
import type {
  CreateTaskInput,
  ListParams,
  Task,
  UpdateTaskInput,
} from "@/lib/types";
import { useDebounce } from "@/lib/useDebounce";
import Navbar from "@/components/Navbar";
import Toolbar from "@/components/Toolbar";
import TaskItem from "@/components/TaskItem";
import TaskFormModal from "@/components/TaskFormModal";
import Pagination from "@/components/Pagination";

const PAGE_SIZE = 8;

export default function TasksPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();

  // List query state.
  const [params, setParams] = useState<ListParams>({
    status: "",
    search: "",
    sortBy: "created_at",
    sortDir: "desc",
    page: 1,
    pageSize: PAGE_SIZE,
  });
  const debouncedSearch = useDebounce(params.search ?? "", 350);

  // Data state.
  const [tasks, setTasks] = useState<Task[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // IDs with an in-flight optimistic mutation (greyed out in the UI).
  const [pendingIds, setPendingIds] = useState<Set<string>>(new Set());

  // Modal state: undefined = closed, null = creating, Task = editing.
  const [modalTask, setModalTask] = useState<Task | null | undefined>(undefined);

  // Guard: bounce unauthenticated visitors to login.
  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  const effectiveParams = useMemo<ListParams>(
    () => ({ ...params, search: debouncedSearch }),
    [params, debouncedSearch]
  );

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api.listTasks(effectiveParams);
      setTasks(res.tasks);
      setTotal(res.total);
      setTotalPages(res.totalPages);
    } catch (err) {
      setError(
        err instanceof ApiClientError
          ? err.message
          : "Could not load tasks. Is the API running?"
      );
    } finally {
      setLoading(false);
    }
  }, [effectiveParams]);

  useEffect(() => {
    if (user) fetchTasks();
  }, [user, fetchTasks]);

  // Updates the query, resetting to page 1 whenever a filter/search/sort changes.
  function updateParams(next: Partial<ListParams>) {
    setParams((prev) => {
      const resetPage = "page" in next ? {} : { page: 1 };
      return { ...prev, ...next, ...resetPage };
    });
  }

  function markPending(id: string, on: boolean) {
    setPendingIds((prev) => {
      const copy = new Set(prev);
      if (on) copy.add(id);
      else copy.delete(id);
      return copy;
    });
  }

  // --- Optimistic mutations ---

  async function toggleComplete(task: Task) {
    const nextStatus = task.status === "done" ? "todo" : "done";
    const prevTasks = tasks;
    // Optimistically flip the status before the server confirms.
    setTasks((ts) =>
      ts.map((t) => (t.id === task.id ? { ...t, status: nextStatus } : t))
    );
    markPending(task.id, true);
    try {
      await api.updateTask(task.id, { status: nextStatus });
    } catch {
      setTasks(prevTasks); // rollback on failure
      setError("Could not update the task. Reverted.");
    } finally {
      markPending(task.id, false);
    }
  }

  async function deleteTask(task: Task) {
    if (!confirm(`Delete "${task.title}"?`)) return;
    const prevTasks = tasks;
    // Optimistically remove from the list.
    setTasks((ts) => ts.filter((t) => t.id !== task.id));
    setTotal((n) => Math.max(0, n - 1));
    markPending(task.id, true);
    try {
      await api.deleteTask(task.id);
    } catch {
      setTasks(prevTasks);
      setTotal((n) => n + 1);
      setError("Could not delete the task. Reverted.");
    } finally {
      markPending(task.id, false);
    }
  }

  async function createTask(input: CreateTaskInput) {
    await api.createTask(input);
    await fetchTasks();
  }

  async function updateTask(id: string, input: UpdateTaskInput) {
    await api.updateTask(id, input);
    await fetchTasks();
  }

  if (authLoading || !user) {
    return <div className="spinner" />;
  }

  return (
    <>
      <Navbar />
      <main className="container">
        <div className="header-row">
          <h1>Your tasks</h1>
          <button className="btn btn-primary" onClick={() => setModalTask(null)}>
            + New task
          </button>
        </div>

        <Toolbar params={params} onChange={updateParams} />

        {error && <div className="banner-error">{error}</div>}

        {loading ? (
          <div className="spinner" />
        ) : tasks.length === 0 ? (
          <div className="state">
            {total === 0 && !params.search && !params.status ? (
              <>
                <p>No tasks yet.</p>
                <button className="btn btn-primary" onClick={() => setModalTask(null)}>
                  Create your first task
                </button>
              </>
            ) : (
              <p>No tasks match your filters.</p>
            )}
          </div>
        ) : (
          <div className="task-list">
            {tasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                pending={pendingIds.has(task.id)}
                onToggleComplete={toggleComplete}
                onEdit={(t) => setModalTask(t)}
                onDelete={deleteTask}
              />
            ))}
          </div>
        )}

        <Pagination
          page={params.page ?? 1}
          totalPages={totalPages}
          total={total}
          onChange={(page) => updateParams({ page })}
        />
      </main>

      {modalTask !== undefined && (
        <TaskFormModal
          task={modalTask}
          onClose={() => setModalTask(undefined)}
          onCreate={createTask}
          onUpdate={updateTask}
        />
      )}
    </>
  );
}
