import React, { useMemo, useState } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useParams } from '@tanstack/react-router';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from '@voxire/utils';
import { HelmetContainer } from '@voxire/components';
import { STATUS_COLORS, useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { monthCells, tasksByLocalDay, agendaDays } from '../../shared/lib/monthGrid';
import { CalendarAgenda } from '../../shared/components/CalendarAgenda';
import { todayIso } from '../../shared/lib/dateUtils';

/**
 * ProjectCalendar - month grid with tasks placed on their due_date.
 * Clicking a task should open its detail (wired via Link in a later pass);
 * for now tasks render as compact chips on the owning day.
 */
/**
 * The grid needs room; below that the month is a list.
 *
 * A min-width test rather than max-width, and it decides which layout is
 * *rendered* rather than which is hidden. Two reasons: a month grid plus a full
 * agenda is a lot of nodes to build and throw away on every render, and with
 * both mounted every task title exists twice in the DOM, which a test cannot
 * disambiguate and a screen reader only avoids because `display: none` happens
 * to prune it.
 */
const WIDE = '(min-width: 768px)';

const ProjectCalendar: React.FC = () => {
  const { t, i18n } = useTranslation();
  const openTask = useOpenTask();
  const isWide = useMediaQuery(WIDE);
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams({ strict: false });

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectLoadFailed,
    refetch: refetchProjects,
  } = useCurrentProject(workspaceId, projectId);
  const { data: apiTasks = [] } = useTasks({ projectId, activeOnly: true, compact: true });

  const [cursor, setCursor] = useState(() => {
    const d = new Date();
    d.setDate(1);
    return d;
  });

  const { cells, weekCount, monthLabel } = useMemo(() => {
    const all = monthCells(cursor);
    return {
      cells: all,
      weekCount: all.length / 7,
      // `undefined` here meant "the browser's locale", which is unrelated to
      // the language the user picked in Flow. The global calendar threads
      // i18n.language through; this now matches it.
      monthLabel: cursor.toLocaleString(i18n.language, {
        month: 'long',
        year: 'numeric',
      }),
    };
  }, [cursor, i18n.language]);

  const tasksByDay = useMemo(() => tasksByLocalDay(apiTasks), [apiTasks]);

  const agenda = useMemo(() => agendaDays(cells, tasksByDay), [cells, tasksByDay]);

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

  // A failed project list leaves `project` undefined exactly as a missing one
  // does, so a 500 used to tell the user their project had gone.
  if (projectLoadFailed) {
    return <ListLoadFailed onRetry={() => void refetchProjects()} />;
  }

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.notFound')}
      </div>
    );
  }

  const today = todayIso();

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

      {/* At 390px a seven-column month gives each day about 55px, which truncates
          a title to two or three characters (rtylr#1215, S-10). */}
      {!isWide && (
        <div className="flex-1 overflow-auto">
          <CalendarAgenda
            entries={agenda}
            todayIso={today}
            taskId={(task) => String(task.id)}
            taskTitle={(task) => task.title}
            taskTone={(task) => STATUS_COLORS[task.status]}
            onOpen={(task) => openTask(String(task.id))}
          />
        </div>
      )}

      {isWide && (
        <div className="flex flex-1 flex-col px-3 pb-3 min-h-0">
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
              gridTemplateRows: `repeat(${weekCount}, minmax(96px, 1fr))`,
              background: 'var(--flow-border)',
            }}
          >
            {cells.map((cell, i) => {
              const d = cell.inMonth ? cell.date : null;
              const dayTasks = d ? (tasksByDay.get(cell.iso) ?? []) : [];
              const isToday = cell.inMonth && cell.iso === today;
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
                        {dayTasks.slice(0, 3).map((t) => {
                          // Tinted by status. Every task on this screen belongs to
                          // the same project, so project.color made all 34 pills
                          // one shade and carried no information (#1087).
                          const tone = STATUS_COLORS[t.status] ?? 'var(--flow-text-muted)';
                          return (
                            <div
                              key={t.id}
                              // dir="auto" so a right-to-left title truncates at
                              // its own end. Left as-is, an Arabic task clipped
                              // with the ellipsis at the start of the phrase.
                              dir="auto"
                              className="text-meta px-1.5 py-0.5 rounded-sm truncate"
                              style={{
                                // See Calendar: the tone is set once and read back,
                                // which also makes it observable in tests.
                                ['--tone' as string]: tone,
                                background: withAlpha('var(--tone)', 13),
                                color: onTint('var(--tone)'),
                                cursor: 'pointer',
                              }}
                              title={t.title}
                              onClick={() => openTask(String(t.id))}
                            >
                              {t.title}
                            </div>
                          );
                        })}
                        {dayTasks.length > 3 && (
                          <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                            {t('flow.calendar.moreCount', { count: dayTasks.length - 3 })}
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
      )}
    </div>
  );
};

export default ProjectCalendar;
