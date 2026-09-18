import React from 'react';
import { Timer } from 'lucide-react';
import { Cycle } from '../stores/flowStore';
import { withAlpha } from '../lib/colorUtils';

interface CycleBadgeProps {
  cycle: Cycle;
  compact?: boolean;
}

const STATUS_COLORS: Record<Cycle['status'], string> = {
  active: 'var(--flow-status-approved)',
  upcoming: 'var(--flow-info-fg)',
  completed: 'var(--flow-neutral-fg)',
};

export const CycleBadge: React.FC<CycleBadgeProps> = ({ cycle, compact }) => {
  const pct = cycle.scope > 0 ? Math.round((cycle.completed / cycle.scope) * 100) : 0;
  const color = STATUS_COLORS[cycle.status];
  return (
    <div
      className="inline-flex items-center gap-1.5 rounded-full"
      style={{
        // withAlpha, not a hex suffix: STATUS_COLORS.active is a token now,
        // and `var(--flow-status-approved)18` is not a colour, so the whole
        // declaration would be dropped and the badge would lose its fill.
        background: withAlpha(color, 9),
        border: `1px solid ${withAlpha(color, 20)}`,
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
