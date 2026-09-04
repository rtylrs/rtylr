import React, { useMemo, useState } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';

/**
 * ProjectCalendar - month grid with tasks placed on their due_date.
 * Clicking a task should open its detail (wired via Link in a later pass);
 * for now tasks render as compact chips on the owning day.
 */
const ProjectCalendar: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams({ strict: false });

  useProjectRealtime(projectId);

  const navigate = useNavigate();
  const { project, isLoading: projectLoading } = useCurrentProject(workspaceId, projectId);
  const { data: apiTasks = [] } = useTasks({ projectId, activeOnly: true, compact: true });

  const [cursor, setCursor] = useState(() => {
    const d = new Date();
    d.setDate(1);
    return d;
  });

  const { weeks, monthLabel } = useMemo(() => {
    const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1);
    const startDay = first.getDay(); // 0=Sun
    const daysInMonth = new Date(cursor.getFullYear(), cursor.getMonth() + 1, 0).getDate();
    const cells: (Date | null)[] = [];
    for (let i = 0; i < startDay; i++) cells.push(null);
    for (let i = 1; i <= daysInMonth; i++) {
      cells.push(new Date(cursor.getFullYear(), cursor.getMonth(), i));
    }
    while (cells.length % 7 !== 0) cells.push(null);

    const rows: (Date | null)[][] = [];
    for (let i = 0; i < cells.length; i += 7) rows.push(cells.slice(i, i + 7));

    return {
      weeks: rows,
      // `undefined` here meant "the browser's locale", which is unrelated to
      // the language the user picked in Flow. The global calendar threads
      // i18n.language through; this now matches it.
      monthLabel: cursor.toLocaleString(i18n.language, {
        month: 'long',
        year: 'numeric',
      }),
    };
  }, [cursor, i18n.language]);

  const tasksByDay = useMemo(() => {
    const map = new Map<string, typeof apiTasks>();
    apiTasks.forEach((t) => {
      if (!t.dueDate) return;
      const key = new Date(t.dueDate).toDateString();
      const bucket = map.get(key) ?? [];
      bucket.push(t);
      map.set(key, bucket);
    });
    return map;
  }, [apiTasks]);

  // Built from Intl rather than hardcoded English. The grid's first column is
  // Sunday, so the names are generated from a known Sunday; which day the week
  // starts on is unchanged, only the language of the labels. Declared here,
  // above the early returns below, because a hook has to run on every render.
  const weekdays = useMemo(() => {
    const formatter = new Intl.DateTimeFormat(i18n.language, { weekday: 'short' });
    const knownSunday = new Date(2026, 0, 4);
    return Array.from({ length: 7 }, (_, offset) =>
      formatter.format(new Date(2026, 0, knownSunday.getDate() + offset))
    );
  }, [i18n.language]);

  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.loading')}
      </div>
    );
  }

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.notFound')}
      </div>
    );
  }

  const today = new Date().toDateString();

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectCalendar.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
      />

      <div className="flex items-center justify-between px-4 py-2">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setCursor((d) => new Date(d.getFullYear(), d.getMonth() - 1, 1))}
            className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
            aria-label={t('flow.calendar.prevMonthLabel')}
          >
            <ChevronLeft size={14} />
          </button>
          <button
            type="button"
            onClick={() => setCursor((d) => new Date(d.getFullYear(), d.getMonth() + 1, 1))}
            className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
            aria-label={t('flow.calendar.nextMonthLabel')}
          >
            <ChevronRight size={14} />
          </button>
          <h2 className="text-body font-semibold" style={{ color: 'var(--flow-text)' }}>
            {monthLabel}
          </h2>
        </div>
        <button
          type="button"
          onClick={() => {
            const d = new Date();
            d.setDate(1);
            setCursor(d);
          }}
          className="text-meta px-2 py-1 rounded-sm hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.calendar.today')}
        </button>
      </div>

      <div className="flex-1 flex flex-col px-3 pb-3 min-h-0">
        <div className="grid grid-cols-7 gap-px shrink-0">
          {weekdays.map((d) => (
            <div
              key={d}
              className="text-meta font-semibold uppercase tracking-wider py-1.5 text-center"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              {d}
            </div>
          ))}
        </div>
        <div
          className="flex-1 grid gap-px overflow-auto rounded-md"
          style={{
            gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
            gridTemplateRows: `repeat(${weeks.length}, minmax(96px, 1fr))`,
            background: 'var(--flow-border)',
          }}
        >
          {weeks.flat().map((d, i) => {
            const key = d?.toDateString() ?? '';
            const dayTasks = d ? (tasksByDay.get(key) ?? []) : [];
            const isToday = key === today;
            return (
              <div
                key={i}
                className="flex flex-col p-1.5 overflow-hidden"
                style={{
                  background: d ? 'var(--flow-card)' : 'var(--flow-surface)',
                }}
              >
                {d && (
                  <>
                    <div
                      className={[
                        'text-meta mb-1 self-end',
                        isToday ? 'font-bold' : 'font-medium',
                      ].join(' ')}
                      style={{
                        color: isToday ? 'var(--accent)' : 'var(--flow-text-dim)',
                      }}
                    >
                      {d.getDate()}
                    </div>
                    <div className="flex flex-col gap-1 min-h-0 overflow-hidden">
                      {dayTasks.slice(0, 3).map((t) => (
                        <div
                          key={t.id}
                          className="text-meta px-1.5 py-0.5 rounded-sm truncate"
                          style={{
                            background: `${project.color}22`,
                            color: project.color,
                            cursor: 'pointer',
                          }}
                          title={t.title}
                          onClick={() => navigate({ to: `/task/${t.id}` as string })}
                        >
                          {t.title}
                        </div>
                      ))}
                      {dayTasks.length > 3 && (
                        <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                          +{dayTasks.length - 3} more
                        </span>
                      )}
                    </div>
                  </>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default ProjectCalendar;
