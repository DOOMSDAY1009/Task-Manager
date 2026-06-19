import type { Task } from "./types";

// Formats an ISO timestamp as a short local date, or "" when absent.
export function formatDate(iso: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// Converts an ISO timestamp to the value format expected by <input type="date">.
export function toDateInput(iso: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toISOString().slice(0, 10);
}

// Converts a yyyy-mm-dd input value to an RFC3339 timestamp (UTC midnight).
export function fromDateInput(value: string): string | null {
  if (!value) return null;
  return new Date(value + "T00:00:00Z").toISOString();
}

// A task is overdue if it has a due date in the past and is not done.
export function isOverdue(task: Task): boolean {
  if (!task.dueDate || task.status === "done") return false;
  return new Date(task.dueDate).getTime() < Date.now();
}

export const STATUS_LABELS: Record<string, string> = {
  todo: "To do",
  in_progress: "In progress",
  done: "Done",
};
