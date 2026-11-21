import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { ApiEpicRollup } from '../services/flowApi';
import { EPIC_PROGRESS_MODES, epicProgress, type EpicProgressMode } from '../lib/taskHierarchy';
import { formatDuration } from '../lib/durationUtils';
import FlowSelect from './FlowSelect';

const MODE_KEY = 'flow:epics:progressMode';

function loadMode(): EpicProgressMode {
  try {
    const raw = localStorage.getItem(MODE_KEY);
    if ((EPIC_PROGRESS_MODES as readonly string[]).includes(raw ?? '')) {
      return raw as EpicProgressMode;
    }
  } catch {
    // Storage unavailable: count is the default.
  }
  return 'count';
}

/**
 * The progress mode, one preference for every epic bar (task panel and
 * dashboard), remembered on this device.
 */
export function useEpicProgressMode(): [EpicProgressMode, (mode: EpicProgressMode) => void] {
  const [mode, setModeState] = useState<EpicProgressMode>(loadMode);
  const setMode = (next: EpicProgressMode): void => {
    setModeState(next);
    try {
      localStorage.setItem(MODE_KEY, next);
    } catch {
      // A preference that does not stick still applies for this visit.
    }
  };
  return [mode, setMode];
}

export const EpicProgressModeSelect: React.FC<{
  mode: EpicProgressMode;
  onChange: (mode: EpicProgressMode) => void;
}> = ({ mode, onChange }) => {
  const { t } = useTranslation();
  return (
    <FlowSelect
      label={t('flow.epics.progressMode')}
      value={mode}
      options={EPIC_PROGRESS_MODES.map((m) => ({ value: m, label: t(`flow.epics.mode.${m}`) }))}
      onChange={(value) => onChange(value as EpicProgressMode)}
      muted
    />
  );
};

/** One epic's progress: a bar and what it measures, in the chosen mode. */
export const EpicProgressBar: React.FC<{
  rollup: ApiEpicRollup;
  mode: EpicProgressMode;
}> = ({ rollup, mode }) => {
  const { t } = useTranslation();
  const progress = epicProgress(rollup, mode);
  const tone = rollup.color || 'var(--accent)';
  const summary =
    mode === 'count'
      ? t('flow.epics.countSummary', { done: progress.done, total: progress.total })
      : progress.total === 0
        ? t('flow.epics.noEstimate')
        : t(mode === 'hours' ? 'flow.epics.hoursSummary' : 'flow.epics.estimateSummary', {
            done: formatDuration(progress.done),
            total: formatDuration(progress.total),
          });
  return (
    <div className="min-w-0">
      <div
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={progress.percent}
        aria-label={t('flow.epics.progressLabel', { name: rollup.title ?? '' })}
        className="h-1.5 rounded-full overflow-hidden"
        style={{ background: 'var(--flow-hover)' }}
      >
        <div
          className="h-full rounded-full"
          style={{
            width: `${progress.percent}%`,
            background: progress.over ? 'var(--flow-danger-fg)' : tone,
          }}
        />
      </div>
      <div
        className="mt-1 flex items-center justify-between gap-2 text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <span className="truncate">{summary}</span>
        <span className="shrink-0 font-mono tabular-nums">{progress.percent}%</span>
      </div>
    </div>
  );
};

export default EpicProgressBar;
