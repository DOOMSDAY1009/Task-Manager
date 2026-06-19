"use client";

interface Props {
  page: number;
  totalPages: number;
  total: number;
  onChange: (page: number) => void;
}

export default function Pagination({ page, totalPages, total, onChange }: Props) {
  if (total === 0) return null;
  return (
    <div className="pagination">
      <button
        className="btn"
        disabled={page <= 1}
        onClick={() => onChange(page - 1)}
      >
        ← Prev
      </button>
      <span className="page-info">
        Page {page} of {Math.max(totalPages, 1)} · {total} task{total === 1 ? "" : "s"}
      </span>
      <button
        className="btn"
        disabled={page >= totalPages}
        onClick={() => onChange(page + 1)}
      >
        Next →
      </button>
    </div>
  );
}
