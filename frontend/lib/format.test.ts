import { describe, it, expect } from "vitest";
import { fromDateInput, toDateInput, isOverdue, formatDate } from "./format";
import type { Task } from "./types";

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: "1",
    userId: "u1",
    title: "t",
    description: "",
    status: "todo",
    priority: "medium",
    dueDate: null,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("date helpers", () => {
  it("round-trips a date through input format", () => {
    const iso = fromDateInput("2025-03-15");
    expect(iso).not.toBeNull();
    expect(toDateInput(iso)).toBe("2025-03-15");
  });

  it("treats empty input as null", () => {
    expect(fromDateInput("")).toBeNull();
    expect(toDateInput(null)).toBe("");
  });

  it("formats null as empty string", () => {
    expect(formatDate(null)).toBe("");
  });
});

describe("isOverdue", () => {
  it("is true for a past due date that is not done", () => {
    expect(isOverdue(makeTask({ dueDate: "2000-01-01T00:00:00Z" }))).toBe(true);
  });

  it("is false when the task is done", () => {
    expect(
      isOverdue(makeTask({ dueDate: "2000-01-01T00:00:00Z", status: "done" }))
    ).toBe(false);
  });

  it("is false when there is no due date", () => {
    expect(isOverdue(makeTask({ dueDate: null }))).toBe(false);
  });

  it("is false for a future due date", () => {
    const future = new Date(Date.now() + 86_400_000).toISOString();
    expect(isOverdue(makeTask({ dueDate: future }))).toBe(false);
  });
});
