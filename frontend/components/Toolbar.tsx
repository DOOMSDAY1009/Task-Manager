"use client";

import type { ListParams } from "@/lib/types";

interface Props {
  params: ListParams;
  onChange: (next: Partial<ListParams>) => void;
}

// Search + status filter + sort controls. Changes flow up so the parent can
// refetch; filters/search/sort all coexist in the same query.
export default function Toolbar({ params, onChange }: Props) {
  return (
    <div className="toolbar">
      <div className="form-group grow">
        <label htmlFor="search">Search</label>
        <input
          id="search"
          type="search"
          placeholder="Search by title…"
          value={params.search ?? ""}
          onChange={(e) => onChange({ search: e.target.value })}
        />
      </div>

      <div className="form-group">
        <label htmlFor="status">Status</label>
        <select
          id="status"
          value={params.status ?? ""}
          onChange={(e) => onChange({ status: e.target.value as ListParams["status"] })}
        >
          <option value="">All</option>
          <option value="todo">To do</option>
          <option value="in_progress">In progress</option>
          <option value="done">Done</option>
        </select>
      </div>

      <div className="form-group">
        <label htmlFor="sortBy">Sort by</label>
        <select
          id="sortBy"
          value={params.sortBy ?? "created_at"}
          onChange={(e) => onChange({ sortBy: e.target.value as ListParams["sortBy"] })}
        >
          <option value="created_at">Created date</option>
          <option value="due_date">Due date</option>
          <option value="priority">Priority</option>
        </select>
      </div>

      <div className="form-group">
        <label htmlFor="sortDir">Order</label>
        <select
          id="sortDir"
          value={params.sortDir ?? "desc"}
          onChange={(e) => onChange({ sortDir: e.target.value as ListParams["sortDir"] })}
        >
          <option value="desc">Desc</option>
          <option value="asc">Asc</option>
        </select>
      </div>
    </div>
  );
}
