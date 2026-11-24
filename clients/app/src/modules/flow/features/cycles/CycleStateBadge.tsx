import React from 'react';
import { useTranslation } from 'react-i18next';
import type { ApiCycle } from '../../shared/services/flowApi';
import { withAlpha } from '../../shared/lib/colorUtils';

const STATE_COLORS: Record<ApiCycle['status'], string> = {
  active: 'var(--flow-status-approved)',
  upcoming: 'var(--flow-info-fg)',
  completed: 'var(--flow-neutral-fg)',
};

/** The cycle's state (Upcoming, Active, Done), for places that already show its name. */
const CycleStateBadge: React.FC<{ status: ApiCycle['status'] }> = ({ status }) => {
  const { t } = useTranslation();
  const color = STATE_COLORS[status] ?? STATE_COLORS.upcoming;
  return (
    <span
      className="inline-flex items-center rounded-full px-2 py-px text-flow-meta font-semibold"
      style={{
        background: withAlpha(color, 9),
        border: `1px solid ${withAlpha(color, 20)}`,
        color,
      }}
    >
      {t(`flow.cycles.state.${status}`)}
    </span>
  );
};

export default CycleStateBadge;
