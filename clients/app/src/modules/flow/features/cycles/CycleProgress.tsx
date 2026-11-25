import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowRight } from 'lucide-react';
import type { ApiCycle } from '../../shared/services/flowApi';
import { formatShortDate } from '../../shared/lib/utils';
import { SummaryCard } from '../projects/components/ProjectSummaryCards';
import { useCycles } from './useCycles';
import { cyclePercent } from './cycleModel';

/** "3 of 8 tasks done", the percentage and a bar. */
export const CycleProgress: React.FC<{ cycle: Pick<ApiCycle, 'scope' | 'completed'> }> = ({
  cycle,
}) => {
  const { t } = useTranslation();
  const percent = cyclePercent(cycle);
  return (
    <div>
      <div className="flex items-center justify-between mb-1.5">
        <span className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.cycles.taskCount', { done: cycle.completed, total: cycle.scope })}
        </span>
        <span className="text-flow-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
          {percent}%
        </span>
      </div>
      <div
        className="h-2 rounded-full overflow-hidden"
        style={{ background: 'var(--flow-hover)' }}
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={percent}
        aria-label={t('flow.cycles.detail.progress')}
      >
        <div
          className="h-full rounded-full"
          style={{ width: `${percent}%`, background: 'var(--flow-action-bg)' }}
        />
      </div>
    </div>
  );
};

/** The project's active cycle and its progress, linking to its page. Nothing when none is active. */
export const ActiveCycleCard: React.FC<{ projectId: string }> = ({ projectId }) => {
  const { t, i18n } = useTranslation();
  const { data: cycles = [] } = useCycles(projectId || undefined);
  const active = cycles.find((c) => c.status === 'active');
  if (!active) return null;
  return (
    <SummaryCard title={t('flow.cycles.activeCycle')}>
      <div className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <Link
            to="/flow/cycles/$cycleId"
            params={{ cycleId: active.id }}
            className="flow-focus text-flow-body font-semibold hover:underline"
            style={{ color: 'var(--flow-text)' }}
          >
            {active.name}
          </Link>
          <span
            className="flex items-center gap-2 text-flow-meta"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {formatShortDate(active.startDate, i18n.language)}
            <ArrowRight size={10} className="rtl:rotate-180" aria-hidden />
            {formatShortDate(active.endDate, i18n.language)}
          </span>
        </div>
        <CycleProgress cycle={active} />
      </div>
    </SummaryCard>
  );
};
