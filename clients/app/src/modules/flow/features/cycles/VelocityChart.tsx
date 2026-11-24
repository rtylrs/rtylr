import React from 'react';
import { useTranslation } from 'react-i18next';
import type { VelocityEntry } from './cycleModel';
import { withAlpha } from '../../shared/lib/colorUtils';

const COMMITTED_COLOR = withAlpha('var(--accent)', 25);

const BAR_AREA = 120;

/**
 * Committed against completed story points for the last completed cycles
 * (T-381). Plain HTML bars so names wrap and the layout mirrors in Arabic.
 */
const VelocityChart: React.FC<{ entries: VelocityEntry[]; average: number }> = ({
  entries,
  average,
}) => {
  const { t } = useTranslation();
  const max = Math.max(1, ...entries.flatMap((e) => [e.committed, e.completed]));
  const height = (value: number) => `${Math.round((value / max) * BAR_AREA)}px`;

  return (
    <figure>
      <div
        className="flex items-end gap-4 overflow-x-auto pb-1"
        role="img"
        aria-label={t('flow.cycles.velocity.summary', { average, count: entries.length })}
      >
        {entries.map((entry) => (
          <div key={entry.id} className="flex flex-col items-center gap-1 min-w-[56px]">
            <div className="flex items-end gap-1" style={{ height: BAR_AREA }}>
              <div
                className="w-4 rounded-t-sm"
                style={{ height: height(entry.committed), background: COMMITTED_COLOR }}
                title={t('flow.cycles.velocity.committedValue', { value: entry.committed })}
              />
              <div
                className="w-4 rounded-t-sm"
                style={{ height: height(entry.completed), background: 'var(--accent)' }}
                title={t('flow.cycles.velocity.completedValue', { value: entry.completed })}
              />
            </div>
            <span className="text-flow-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
              {entry.completed}/{entry.committed}
            </span>
            <span
              className="text-flow-meta max-w-[80px] truncate"
              style={{ color: 'var(--flow-text-dim)' }}
              title={entry.name}
            >
              {entry.name}
            </span>
          </div>
        ))}
      </div>
      <figcaption
        className="mt-2 flex flex-wrap items-center gap-4 text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm" style={{ background: COMMITTED_COLOR }} />
          {t('flow.cycles.velocity.committed')}
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm" style={{ background: 'var(--accent)' }} />
          {t('flow.cycles.velocity.completed')}
        </span>
        <span>{t('flow.cycles.velocity.average', { average })}</span>
      </figcaption>
    </figure>
  );
};

export default VelocityChart;
