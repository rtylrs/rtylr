import React from 'react';
import { Timer } from 'lucide-react';
import { Cycle } from '../stores/flowStore';

interface CycleBadgeProps {
  cycle: Cycle;
  compact?: boolean;
}

const STATUS_COLORS: Record<Cycle['status'], string> = {
  active: '#22c55e',
  upcoming: '#3b82f6',
  completed: '#6b7280',
};

export const CycleBadge: React.FC<CycleBadgeProps> = ({ cycle, compact }) => {
  const pct = cycle.scope > 0 ? Math.round((cycle.completed / cycle.scope) * 100) : 0;
  const color = STATUS_COLORS[cycle.status];
  return (
    <div
      className="inline-flex items-center gap-1.5 rounded-full"
      style={{
        background: `${color}18`,
        border: `1px solid ${color}33`,
        color,
        padding: compact ? '1px 6px' : '2px 10px',
        fontSize: 'var(--text-meta)',
      }}
    >
      <Timer size={compact ? 10 : 12} />
      <span className="font-semibold">{cycle.name}</span>
      {!compact && (
        <>
          <span style={{ opacity: 0.5 }}>·</span>
          <span>{pct}%</span>
        </>
      )}
    </div>
  );
};

export default CycleBadge;
