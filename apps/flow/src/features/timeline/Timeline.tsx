import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore, type Project, type Milestone } from '../../shared/stores/flowStore';
import { useProjects } from '../projects/useProjects';
import { useMilestones } from '../milestones/useMilestones';
import Breadcrumbs from '../../shared/components/Breadcrumbs';

function parseDate(d: string): number {
  return new Date(d).getTime();
}

const DAY_MS = 24 * 60 * 60 * 1000;

const Timeline: React.FC = () => {
  const { t, i18n } = useTranslation();
  const storeProjects = useFlowStore((s) => s.projects);
  const storeMilestones = useFlowStore((s) => s.milestones);
  const { data: apiProjects } = useProjects();
  const { data: apiMilestones } = useMilestones();
  const projects = (apiProjects ?? storeProjects) as Project[];
  const milestones = (apiMilestones ?? storeMilestones) as Milestone[];

  const { start, end, days } = useMemo(() => {
    const now = new Date();
    const s = new Date(now.getFullYear(), now.getMonth() - 1, 1).getTime();
    const e = new Date(now.getFullYear(), now.getMonth() + 3, 0).getTime();
    const d: { label: string; time: number; isMonth: boolean }[] = [];
    for (let t = s; t <= e; t += DAY_MS) {
      const dt = new Date(t);
      d.push({
        label:
          dt.getDate() === 1
            ? dt.toLocaleString(i18n.language, { month: 'short' })
            : `${dt.getDate()}`,
        time: t,
        isMonth: dt.getDate() === 1,
      });
    }
    return { start: s, end: e, days: d };
  }, [i18n.language]);

  const total = end - start;

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.shell.timeline')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <Breadcrumbs items={[{ label: t('flow.shell.timeline') }]} />
        <h1 className="text-title font-semibold mt-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.shell.timeline')}
        </h1>
      </div>
      <div className="flex-1 overflow-auto">
        <div className="min-w-[1200px]">
          {/* Header row with dates */}
          <div
            className="sticky top-0 z-10 flex"
            style={{
              background: 'var(--flow-bg)',
              borderBottom: '1px solid var(--flow-border)',
            }}
          >
            <div
              className="w-[192px] shrink-0 px-4 py-2 text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-dim)', borderRight: '1px solid var(--flow-border)' }}
            >
              {t('flow.timeline.projectColumn')}
            </div>
            <div className="flex-1 flex text-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {days
                .filter((d, i) => i % 3 === 0)
                .map((d) => (
                  <div
                    key={d.time}
                    className="flex-1 min-w-[24px] text-center py-2"
                    style={{
                      borderLeft: d.isMonth ? '1px solid var(--flow-border)' : undefined,
                      fontWeight: d.isMonth ? 600 : 400,
                    }}
                  >
                    {d.label}
                  </div>
                ))}
            </div>
          </div>

          {/* Rows */}
          {projects.map((p) => {
            const s = parseDate(p.startDate);
            const e = parseDate(p.targetDate);
            const leftPct = Math.max(0, ((s - start) / total) * 100);
            const widthPct = Math.max(2, ((e - s) / total) * 100);
            const pMilestones = milestones.filter((m) => m.projectId === p.id);
            return (
              <div
                key={p.id}
                className="flex"
                style={{ borderBottom: '1px solid var(--flow-border)' }}
              >
                <div
                  className="w-[192px] shrink-0 px-4 py-3 flex items-center gap-2"
                  style={{ borderRight: '1px solid var(--flow-border)' }}
                >
                  <span className="h-2 w-2 rounded-full" style={{ background: p.color }} />
                  <span className="text-meta font-medium" style={{ color: 'var(--flow-text)' }}>
                    {p.name}
                  </span>
                </div>
                <div className="flex-1 relative py-3">
                  <div
                    className="absolute h-6 rounded-md flex items-center px-2"
                    style={{
                      left: `${leftPct}%`,
                      width: `${widthPct}%`,
                      top: 8,
                      background: `${p.color}33`,
                      border: `1px solid ${p.color}66`,
                    }}
                  >
                    <span className="text-meta font-semibold truncate" style={{ color: p.color }}>
                      {p.progress}%
                    </span>
                  </div>
                  {pMilestones.map((m) => {
                    const mt = parseDate(m.targetDate);
                    const mLeft = ((mt - start) / total) * 100;
                    return (
                      <div
                        key={m.id}
                        title={m.name}
                        className="absolute"
                        style={{
                          left: `${mLeft}%`,
                          top: 8,
                          width: 0,
                          height: 24,
                        }}
                      >
                        <div
                          className="absolute -translate-x-1/2 -top-0.5 h-6 w-6 flex items-center justify-center rotate-45"
                          style={{
                            background: 'var(--flow-warn-fg)',
                            borderRadius: 3,
                          }}
                        />
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default Timeline;
