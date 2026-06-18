import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationControlsProps {
  page: number;
  hasMore: boolean;
  itemCount?: number;
  onPageChange: (page: number) => void;
}

const PaginationControls: React.FC<PaginationControlsProps> = ({
  page,
  hasMore,
  itemCount,
  onPageChange,
}) => {
  if (page <= 1 && !hasMore) return null;

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 px-1 py-3">
      <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
        Page {page}
        {itemCount !== undefined ? ` · ${itemCount} shown` : ''}
      </span>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={page <= 1}
          className="flex items-center gap-1 rounded-md px-3 py-1.5 text-xs font-medium disabled:opacity-30"
          style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
        >
          <ChevronLeft size={12} />
          Prev
        </button>
        <button
          type="button"
          onClick={() => onPageChange(page + 1)}
          disabled={!hasMore}
          className="flex items-center gap-1 rounded-md px-3 py-1.5 text-xs font-medium disabled:opacity-30"
          style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
        >
          Next
          <ChevronRight size={12} />
        </button>
      </div>
    </div>
  );
};

export default PaginationControls;
