import React from 'react';
import { Timer } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Cycle } from '../stores/flowStore';
import { withAlpha } from '../lib/colorUtils';

interface CycleBadgeProps {
  cycle: Pick<Cycle, 'name' | 'status' | 'scope' | 'completed'>;
  compact?: boolean;
  /** Show the state (Upcoming, Active, Completed) instead of the name, where the name is already shown. */
  showState?: boolean;
}

const STATUS_COLORS: Record<Cycle['status'], string> = {
  active: 'var(--flow-status-approved)',
  upcoming: 'var(--flow-info-fg)',
  completed: 'var(--flow-neutral-fg)',
};

export const CycleBadge: React.FC<CycleBadgeProps> = ({ cycle, compact, showState }) => {
  const { t } = useTranslation();
  const pct = cycle.scope > 0 ? Math.round((cycle.completed / cycle.scope) * 100) : 0;
  const color = STATUS_COLORS[cycle.status] ?? STATUS_COLORS.upcoming;
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
        fontSize: 'var(--text-flow-meta)',
      }}
    >
      <Timer size={compact ? 10 : 12} aria-hidden />
      <span className="font-semibold">
        {showState ? t(`flow.backlog.status.${cycle.status}`) : cycle.name}
      </span>
      {!compact && !showState && (
        <>
          <span style={{ opacity: 0.5 }}>·</span>
          <span>{pct}%</span>
        </>
      )}
    </div>
  );
};

export default CycleBadge;
