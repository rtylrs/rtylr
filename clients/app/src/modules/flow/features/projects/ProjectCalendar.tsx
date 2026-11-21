import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import React, { useMemo, useState } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from '@voxire/utils';
import { HelmetContainer } from '@voxire/components';
import { STATUS_COLORS, useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { tasksByLocalDay, agendaDays } from '../../shared/lib/monthGrid';
import { CalendarAgenda } from '../../shared/components/CalendarAgenda';
import { todayIso } from '../../shared/lib/dateUtils';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCanEditTask } from '../../shared/lib/currentUser';
import type { ApiTask } from '../../shared/services/flowApi';
import FilterBar from '../../shared/filters/FilterBar';
import { useProjectTaskFilters } from './useProjectTaskFilters';
import { useCalendarView } from '../calendar/useCalendarView';
import { useCalendarScheduling } from '../calendar/useCalendarScheduling';
import { useCalendarPlanning } from '../calendar/useCalendarPlanning';
import { PlanningGrid } from '../calendar/PlanningGrid';
import { UnscheduledPanel, needsScheduling } from '../calendar/UnscheduledPanel';
import { CalendarControls } from '../calendar/CalendarControls';

/**
 * ProjectCalendar - the project's tasks by date, as a month or a week.
 *
 * A task runs from its start date to its due date; the project's cycles and
 * milestones and the team's approved leave sit behind it; open tasks with no
 * dates wait in the Unscheduled panel to be dragged onto a day.
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

/** Room for the Unscheduled panel beside the grid; below it the panel starts closed. */
const LARGE = '(min-width: 1024px)';

const ProjectCalendar: React.FC = () => {
  const { t, i18n } = useTranslation();
  const openTask = useOpenTask();
  const isWide = useMediaQuery(WIDE);
  const isLarge = useMediaQuery(LARGE);
  const { projectId = '' } = useParams({ strict: false });
  const canEdit = useCanEditTask(projectId);
  // null until the user chooses, so the default can follow the width.
  const [panelChoice, setPanelChoice] = useState<boolean | null>(null);
  const panelOpen = panelChoice ?? isLarge;

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectMetadataFailed,
    refetch: refetchProjectMetadata,
  } = useCurrentProject(projectId);
  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: apiTasks = [] } = tasksQuery;
  const tasksLoadFailed = hasQueryFailed(tasksQuery);
  const tasksLoading = tasksQuery.isPending && !tasksLoadFailed;
  const projectLoadFailed = projectMetadataFailed || tasksLoadFailed;
  const refetchProjects = () => {
    void refetchProjectMetadata();
    void tasksQuery.refetch();
  };

  const calendar = useCalendarView(i18n.language);
  const { view, cells, from, to } = calendar;

  // Same filters as the board and list. A month cell fits three pills, so the
  // calendar always hides non-matching tasks rather than fading them.
  const mappedTasks = useMemo(() => apiTasks.map(apiTaskToTask), [apiTasks]);
  const projectFilters = useProjectTaskFilters(projectId, mappedTasks);
  const { matchingIds } = projectFilters;
  const calendarTasks = useMemo(
    () =>
      matchingIds === null ? apiTasks : apiTasks.filter((task) => matchingIds.has(String(task.id))),
    [apiTasks, matchingIds]
  );
  const scheduling = useCalendarScheduling(calendarTasks);
  const shownTasks = scheduling.tasks;
  const tasksByDay = useMemo(() => tasksByLocalDay(shownTasks), [shownTasks]);
  const agenda = useMemo(() => agendaDays(cells, tasksByDay), [cells, tasksByDay]);
  const unscheduledCount = useMemo(
    () => calendarTasks.filter(needsScheduling).length,
    [calendarTasks]
  );
  const cycleProjectIds = useMemo(() => (projectId ? [projectId] : []), [projectId]);
  const planning = useCalendarPlanning({
    cycleProjectIds,
    milestoneProjectId: projectId,
    from,
    to,
  });

  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.loading')}
      </div>
    );
  }

  if (tasksLoading) {
    return (
      <div className="p-6" role="status" style={{ color: 'var(--flow-text-muted)' }}>
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
  // Tinted by status. Every task on this screen belongs to the same project, so
  // project.color made all 34 pills one shade and carried no information (#1087).
  const statusTone = (task: ApiTask): string =>
    STATUS_COLORS[task.status] ?? 'var(--flow-text-muted)';
  const panel = (layout: 'side' | 'stacked') => (
    <UnscheduledPanel
      tasks={shownTasks}
      tone={statusTone}
      canEdit={() => canEdit}
      onOpen={(task) => openTask(String(task.id))}
      onClose={() => setPanelChoice(false)}
      scheduling={scheduling}
      layout={layout}
    />
  );

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectCalendar.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
      />

      <FilterBar
        state={projectFilters.state}
        currentUserId={projectFilters.currentUserId}
        fields={projectFilters.fields}
        people={projectFilters.people}
        visible={matchingIds?.size ?? apiTasks.length}
        total={apiTasks.length}
        allowDim={false}
      />

      <div className="px-4 py-2">
        <CalendarControls
          view={view}
          onView={calendar.setView}
          label={calendar.label}
          onStep={calendar.step}
          onToday={calendar.goToday}
          unscheduledCount={unscheduledCount}
          unscheduledOpen={panelOpen}
          onToggleUnscheduled={() => setPanelChoice(!panelOpen)}
        />
      </div>

      {/* At 390px a seven-column month gives each day about 55px, which truncates
          a title to two or three characters (rtylr#1215, S-10). */}
      {!isWide && (
        <div className="flex-1 overflow-auto">
          {panelOpen && panel('stacked')}
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
        <div className="flex flex-1 min-h-0">
          <div className="flex flex-1 min-w-0 flex-col px-3 pb-3">
            <PlanningGrid
              cells={cells}
              view={view}
              tasks={shownTasks}
              cycles={planning.cycles}
              milestones={planning.milestones}
              leave={planning.leave}
              today={today}
              tone={statusTone}
              canEdit={() => canEdit}
              onOpen={(task) => openTask(String(task.id))}
              scheduling={scheduling}
            />
          </div>
          {panelOpen && panel('side')}
        </div>
      )}
    </div>
  );
};

export default ProjectCalendar;
