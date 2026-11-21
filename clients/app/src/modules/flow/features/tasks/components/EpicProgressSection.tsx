import React from 'react';
import { useTranslation } from 'react-i18next';
import { BarChart3 } from 'lucide-react';
import { useEpicRollups } from '../useTasks';
import EpicProgressBar, {
  EpicProgressModeSelect,
  useEpicProgressMode,
} from '../../../shared/components/EpicProgressBar';

/** An epic's progress over its tasks, at the top of the epic's panel. */
const EpicProgressSection: React.FC<{ epicId: string; projectId: string }> = ({
  epicId,
  projectId,
}) => {
  const { t } = useTranslation();
  const [mode, setMode] = useEpicProgressMode();
  const { data: rollups } = useEpicRollups(projectId);
  const rollup = rollups?.find((r) => String(r.epicId) === epicId);
  if (!rollup) return null;
  return (
    <section className="mb-8" aria-label={t('flow.epics.progress')}>
      <div className="mb-3 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <BarChart3 size={12} style={{ color: 'var(--flow-text-muted)' }} />
          <h3
            className="text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.epics.progress')}
          </h3>
        </div>
        <div className="w-[140px]">
          <EpicProgressModeSelect mode={mode} onChange={setMode} />
        </div>
      </div>
      <EpicProgressBar rollup={rollup} mode={mode} />
    </section>
  );
};

export default EpicProgressSection;
