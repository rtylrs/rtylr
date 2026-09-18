import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationControlsProps {
  page: number;
  hasMore: boolean;
  itemCount?: number;
  pageSize?: number;
  total?: number;
  onPageChange: (page: number) => void;
}

const PaginationControls: React.FC<PaginationControlsProps> = ({
  page,
  hasMore,
  itemCount,
  pageSize,
  total,
  onPageChange,
}) => {
  if (page <= 1 && !hasMore) return null;

  const hasTotal = typeof total === 'number' && Number.isFinite(total) && total >= 0;
  const hasPageSize = typeof pageSize === 'number' && Number.isFinite(pageSize) && pageSize > 0;
  const start = hasPageSize && (itemCount ?? 0) > 0 ? (page - 1) * pageSize + 1 : undefined;
  const end =
    start !== undefined && itemCount !== undefined ? start + Math.max(0, itemCount - 1) : undefined;
  const pageCount = hasTotal && hasPageSize ? Math.max(1, Math.ceil(total / pageSize)) : undefined;

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 px-1 py-3">
      <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
        Page {page}
        {pageCount !== undefined ? ` of ${pageCount}` : ''}
        {start !== undefined && end !== undefined
          ? ` · ${start}-${end}${hasTotal ? ` of ${total}` : ''}`
          : itemCount !== undefined
            ? ` · ${itemCount} shown`
            : ''}
      </span>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={page <= 1}
          className="flex items-center gap-1 rounded-md px-3 py-1.5 text-xs font-medium disabled:opacity-30"
          style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
        >
          <ChevronLeft size={12} className="erp-directional-icon" />
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
          <ChevronRight size={12} className="erp-directional-icon" />
        </button>
      </div>
    </div>
  );
};

export default PaginationControls;
