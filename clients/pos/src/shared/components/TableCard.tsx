import React from 'react';
import { formatCurrency } from '../utils/format';

export type TableStatus = 'available' | 'seated' | 'ordered' | 'paying' | 'dirty';

interface TableCardProps {
  number: string;
  seats: number;
  status: TableStatus;
  guests?: number;
  total?: number;
  server?: string;
  onClick?: () => void;
}

const STATUS_COLORS: Record<TableStatus, string> = {
  available: 'bg-emerald-500/15 border-emerald-500/30 text-emerald-300',
  seated: 'bg-amber-500/15 border-amber-500/30 text-amber-300',
  ordered: 'bg-blue-500/15 border-blue-500/30 text-blue-300',
  paying: 'bg-purple-500/15 border-purple-500/30 text-purple-300',
  dirty: 'bg-red-500/15 border-red-500/30 text-red-300',
};

// Memoized grid tile so floor-plan re-renders skip unchanged tables.
const TableCard = React.memo<TableCardProps>(function TableCard({
  number,
  seats,
  status,
  guests,
  total,
  server,
  onClick,
}) {
  return (
    <button
      onClick={onClick}
      className={`text-left p-4 rounded-xl border-2 ${STATUS_COLORS[status]} hover:scale-[1.02] transition active:scale-[0.98] min-h-[7rem]`}
    >
      <div className="flex items-start justify-between">
        <span className="text-xl font-bold text-pos-text">#{number}</span>
        <span className="text-[10px] uppercase tracking-wider font-semibold">{status}</span>
      </div>
      <div className="mt-2 text-xs text-pos-muted space-y-0.5">
        <div>
          {guests ?? 0}/{seats} guests
        </div>
        {server && <div>Server: {server}</div>}
        {total != null && total > 0 && (
          <div className="text-pos-text font-semibold">{formatCurrency(total)}</div>
        )}
      </div>
    </button>
  );
});

export default TableCard;
