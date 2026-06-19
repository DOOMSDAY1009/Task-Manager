"use client";

import { useState } from "react";
import type {
  CreateTaskInput,
  Task,
  TaskPriority,
  TaskStatus,
  UpdateTaskInput,
} from "@/lib/types";
import { fromDateInput, toDateInput } from "@/lib/format";
import { ApiClientError } from "@/lib/api";

interface Props {
  // When editing, the existing task; when creating, null.
  task: Task | null;
  onClose: () => void;
  onCreate: (input: CreateTaskInput) => Promise<void>;
  onUpdate: (id: string, input: UpdateTaskInput) => Promise<void>;
}

export default function TaskFormModal({ task, onClose, onCreate, onUpdate }: Props) {
  const editing = task !== null;

  const [title, setTitle] = useState(task?.title ?? "");
  const [description, setDescription] = useState(task?.description ?? "");
  const [status, setStatus] = useState<TaskStatus>(task?.status ?? "todo");
  const [priority, setPriority] = useState<TaskPriority>(task?.priority ?? "medium");
  const [dueDate, setDueDate] = useState(toDateInput(task?.dueDate ?? null));

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [banner, setBanner] = useState("");
  const [saving, setSaving] = useState(false);

  function validate(): boolean {
    const errs: Record<string, string> = {};
    if (!title.trim()) errs.title = "Title is required";
    else if (title.trim().length > 200) errs.title = "Title is too long (max 200)";
    if (description.length > 5000) errs.description = "Description is too long";
    setErrors(errs);
    return Object.keys(errs).length === 0;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBanner("");
    if (!validate()) return;

    setSaving(true);
    try {
      if (editing && task) {
        await onUpdate(task.id, {
          title: title.trim(),
          description,
          status,
          priority,
          // Send the date when set; explicitly clear it when emptied.
          ...(dueDate
            ? { dueDate: fromDateInput(dueDate) }
            : { clearDueDate: true }),
        });
      } else {
        await onCreate({
          title: title.trim(),
          description,
          status,
          priority,
          dueDate: fromDateInput(dueDate),
        });
      }
      onClose();
    } catch (err) {
      if (err instanceof ApiClientError) {
        setBanner(err.message);
        if (err.fields) setErrors(err.fields);
      } else {
        setBanner("Could not save the task. Please try again.");
      }
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <div
        className="modal card"
        role="dialog"
        aria-modal="true"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <h2>{editing ? "Edit task" : "New task"}</h2>
        <form onSubmit={handleSubmit} noValidate>
          {banner && <div className="banner-error">{banner}</div>}

          <div className="form-group">
            <label htmlFor="title">Title</label>
            <input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="What needs doing?"
              autoFocus
            />
            {errors.title && <div className="field-error">{errors.title}</div>}
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional details…"
            />
            {errors.description && (
              <div className="field-error">{errors.description}</div>
            )}
          </div>

          <div className="row">
            <div className="form-group">
              <label htmlFor="status">Status</label>
              <select
                id="status"
                value={status}
                onChange={(e) => setStatus(e.target.value as TaskStatus)}
              >
                <option value="todo">To do</option>
                <option value="in_progress">In progress</option>
                <option value="done">Done</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="priority">Priority</label>
              <select
                id="priority"
                value={priority}
                onChange={(e) => setPriority(e.target.value as TaskPriority)}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="dueDate">Due date</label>
            <input
              id="dueDate"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
            />
          </div>

          <div className="modal-actions">
            <button type="button" className="btn" onClick={onClose} disabled={saving}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? "Saving…" : editing ? "Save changes" : "Create task"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
