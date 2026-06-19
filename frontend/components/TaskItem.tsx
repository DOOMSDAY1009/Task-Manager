"use client";

import type { Task } from "@/lib/types";
import { formatDate, isOverdue, STATUS_LABELS } from "@/lib/format";

interface Props {
  task: Task;
  pending?: boolean;
  onToggleComplete: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
}

export default function TaskItem({
  task,
  pending,
  onToggleComplete,
  onEdit,
  onDelete,
}: Props) {
  const done = task.status === "done";
  const overdue = isOverdue(task);

  return (
    <div className={`task-item${pending ? " pending" : ""}`}>
      <input
        type="checkbox"
        className="check"
        checked={done}
        onChange={() => onToggleComplete(task)}
        aria-label={done ? "Mark as not done" : "Mark as done"}
      />

      <div className="task-main">
        <div className={`task-title${done ? " done" : ""}`}>{task.title}</div>
        {task.description && <div className="task-desc">{task.description}</div>}

        <div className="task-meta">
          <span className={`badge priority-${task.priority}`}>
            {task.priority} priority
          </span>
          <span className={`badge status-${task.status}`}>
            {STATUS_LABELS[task.status]}
          </span>
          {task.dueDate && (
            <span className={`badge due${overdue ? " overdue" : ""}`}>
              {overdue ? "Overdue: " : "Due "}
              {formatDate(task.dueDate)}
            </span>
          )}
        </div>
      </div>

      <div className="task-actions">
        <button className="btn btn-icon" onClick={() => onEdit(task)} title="Edit">
          ✏️
        </button>
        <button
          className="btn btn-danger btn-icon"
          onClick={() => onDelete(task)}
          title="Delete"
        >
          🗑️
        </button>
      </div>
    </div>
  );
}
