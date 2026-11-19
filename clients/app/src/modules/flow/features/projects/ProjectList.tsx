import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import React, { useCallback, useMemo, useRef, useState, useEffect } from 'react';
import { useParams } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { Plus, ChevronDown } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore, Task } from '../../shared/stores/flowStore';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { cn } from '../../shared/lib/utils';
import { todayIso } from '../../shared/lib/dateUtils';
import { useTasks } from '../tasks/useTasks';
import { useLabels } from '../tasks/useLabels';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCurrentProject } from './useCurrentProject';
import { useProjectStatuses } from './useStatuses';
import { defaultStatus, displayStatusName, sortStatuses, statusIdOf } from './statusUtils';
import { DUE_LANES, PRIORITY_LABEL_KEY, PRIORITY_ORDER, laneKeyFor } from './taskGroups';
import FlowSelect from '../../shared/components/FlowSelect';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { useCanCreateTask } from '../../shared/lib/currentUser';
import ProjectViewTabs from './ProjectViewTabs';
import type { ApiStatus } from '../../shared/services/flowApi';
import { VirtualList } from '../../../../shared/components/VirtualList';
import FilterBar from '../../shared/filters/FilterBar';
import { useProjectTaskFilters } from './useProjectTaskFilters';
import { projectFiltersForView, useSaveFiltersAsView } from '../views/useSaveFiltersAsView';
import { TaskRow } from './list/ListRow';
import { ListHeader } from './list/ListHeader';
import { ExportCsvButton, ListColumnPicker } from './list/ListToolbar';
import { ListGridContext, useListGrid, type ListGridState } from './list/listGridContext';
import { handleGridKeyDown, GRID_CELL, GRID_ROW } from './list/gridKeyboard';
import { useListPrefs } from './list/useListPrefs';
import { nextSort, rowMinWidth, sortListTasks, type ListColumnId } from './list/listColumns';
import { csvFileName, downloadCsv, listTasksToCsv } from './list/listCsv';

export { TaskRow };

export type ListGroupKey = 'status' | 'assignee' | 'priority' | 'due' | 'none';

const LIST_GROUP_KEYS: ListGroupKey[] = ['status', 'assignee', 'priority', 'due', 'none'];

const listGroupStorageKey = (projectId: string): string => `flow:list:groupBy:${projectId}`;

export function loadListGroupBy(projectId: string): ListGroupKey {
  try {
    const raw = localStorage.getItem(listGroupStorageKey(projectId));
    if (LIST_GROUP_KEYS.some((key) => key === raw)) return raw as ListGroupKey;
  } catch {
    // Storage unavailable (private mode, blocked site data): use the default.
  }
  return 'status';
}

function saveListGroupBy(projectId: string, key: ListGroupKey): void {
  try {
    localStorage.setItem(listGroupStorageKey(projectId), key);
  } catch {
    // A preference that cannot be stored still applies for this visit.
  }
}

export interface ListGroup {
  id: string;
  label: string;
  color: string;
  tasks: Task[];
  /** The status a task added from this group starts on. */
  statusId?: string;
}

/**
 * Splits the list into groups. Only groups holding tasks are returned, and a
 * task whose status is not in the project's set joins the status with the same
 * built-in value (else the default) rather than an "Other" bucket.
 */
export function buildListGroups(
  tasks: Task[],
  statuses: readonly ApiStatus[],
  groupBy: ListGroupKey,
  t: (key: string) => string
): ListGroup[] {
  const muted = 'var(--flow-text-muted)';
  if (groupBy === 'none') {
    return tasks.length ? [{ id: '_all', label: '', color: muted, tasks }] : [];
  }
  if (groupBy === 'status') {
    const ordered = sortStatuses(statuses);
    const ids = new Set(ordered.map((s) => s.id));
    const fallback = defaultStatus(ordered);
    const homeOf = (task: Task): string | undefined => {
      const own = statusIdOf(task);
      if (own && ids.has(own)) return own;
      return (ordered.find((s) => s.legacyStatus === task.status) ?? fallback)?.id;
    };
    return ordered
      .map((s) => ({
        id: s.id,
        label: displayStatusName(s.name, t),
        color: s.color,
        statusId: s.id,
        tasks: tasks.filter((task) => homeOf(task) === s.id),
      }))
      .filter((group) => group.tasks.length > 0);
  }
  if (groupBy === 'priority') {
    return PRIORITY_ORDER.map((p) => ({
      id: p,
      label: t(PRIORITY_LABEL_KEY[p]),
      color: muted,
      tasks: tasks.filter((task) => task.priority === p),
    })).filter((group) => group.tasks.length > 0);
  }
  if (groupBy === 'due') {
    return DUE_LANES.map((lane) => ({
      id: lane.key,
      label: t(lane.labelKey),
      color: muted,
      tasks: tasks.filter((task) => laneKeyFor(task, 'due') === lane.key),
    })).filter((group) => group.tasks.length > 0);
  }
  // Assignee: people in first-seen order, names never translated.
  const byAssignee = new Map<string, Task[]>();
  for (const task of tasks) {
    const name = task.assignee || '';
    byAssignee.set(name, [...(byAssignee.get(name) ?? []), task]);
  }
  return [...byAssignee.entries()].map(([name, groupTasks]) => ({
    id: name || '_unassigned',
    label: name || t('flow.quickCreate.unassigned'),
    color: muted,
    tasks: groupTasks,
  }));
}

/**
 * One group of the grid: a header row that folds it, its task rows, and a row
 * to add a task into it. Rows window once the group gets long (VirtualList).
 */
export const TaskGroup: React.FC<{
  group: ListGroup;
  /** False for the ungrouped list, which has no header. */
  showHeader?: boolean;
  /** Rows to fade while "Dim instead of hide" is on. */
  dimmedIds?: ReadonlySet<string> | null;
}> = ({ group, showHeader = true, dimmedIds = null }) => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);
  const { projectId } = useParams({ strict: false });
  const canCreate = useCanCreateTask(projectId);
  const open = !showHeader || !collapsed;
  // The Done box and the More button flank the picked columns.
  const span = useListGrid().columns.length + 2;

  return (
    <div role="rowgroup" className="mb-2">
      {showHeader && (
        <div role="row" {...{ [GRID_ROW]: '' }}>
          <div role="gridcell" aria-colspan={span}>
            <button
              type="button"
              {...{ [GRID_CELL]: '' }}
              onClick={() => setCollapsed((c) => !c)}
              aria-expanded={!collapsed}
              className="flow-focus-inset w-full flex items-center gap-2.5 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors"
            >
              <ChevronDown
                size={13}
                className={cn('transition-transform', collapsed ? '-rotate-90 rtl:rotate-90' : '')}
                style={{ color: 'var(--flow-text-dim)' }}
              />
              <span className="h-2 w-2 rounded-full" style={{ background: group.color }} />
              <span className="text-flow-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
                {group.label}
              </span>
              <span
                className="text-flow-meta ms-1 tabular-nums"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {group.tasks.length}
              </span>
            </button>
          </div>
        </div>
      )}

      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.18 }}
            style={{ overflow: 'hidden' }}
          >
            <VirtualList
              count={group.tasks.length}
              estimateSize={41}
              renderRow={(index, rowProps) => (
                <TaskRow
                  key={group.tasks[index].id}
                  rowProps={rowProps}
                  task={group.tasks[index]}
                  dimmed={dimmedIds?.has(group.tasks[index].id) ?? false}
                />
              )}
            />

            {canCreate && (
              <div role="row" {...{ [GRID_ROW]: '' }}>
                <div role="gridcell" aria-colspan={span}>
                  <button
                    type="button"
                    {...{ [GRID_CELL]: '' }}
                    className="flow-focus-inset w-full flex items-center gap-3 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors text-start"
                    style={{
                      borderBottom: '1px solid var(--flow-border)',
                      color: 'var(--flow-text-muted)',
                    }}
                    onClick={() => useFlowStore.getState().setQuickCreateOpen(true, group.statusId)}
                  >
                    <Plus size={14} />
                    <span className="text-flow-body">{t('flow.shell.newTask')}</span>
                  </button>
                </div>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

const ProjectList: React.FC = () => {
  const { t } = useTranslation();
  const { projectId = '' } = useParams({ strict: false });
  const [groupBy, setGroupByState] = useState<ListGroupKey>('status');

  useEffect(() => {
    if (projectId) setGroupByState(loadListGroupBy(projectId));
  }, [projectId]);

  const setGroupBy = (key: ListGroupKey): void => {
    setGroupByState(key);
    if (projectId) saveListGroupBy(projectId, key);
  };

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectMetadataFailed,
    refetch: refetchProjectMetadata,
  } = useCurrentProject(projectId);
  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: apiTasks } = tasksQuery;
  const statusesQuery = useProjectStatuses(projectId);
  const projectLoadFailed =
    projectMetadataFailed || hasQueryFailed(tasksQuery) || hasQueryFailed(statusesQuery);
  const refetchProjects = () => {
    void refetchProjectMetadata();
    void tasksQuery.refetch();
    void statusesQuery.refetch();
  };

  const projectTasks = useMemo(
    () => (apiTasks ?? []).map(apiTaskToTask).filter((task) => !task.archived),
    [apiTasks]
  );
  // Same filters as the board: one set per project, in the URL and remembered.
  const projectFilters = useProjectTaskFilters(projectId, projectTasks);
  const { shownTasks, dimmedIds, matchingIds } = projectFilters;
  const saveAsView = useSaveFiltersAsView();

  // Columns, widths and sort: per user and project, in this browser.
  const [prefs, setPrefs] = useListPrefs(projectId);
  const labelsShown = prefs.columns.includes('labels') || prefs.sort?.column === 'labels';
  const { data: labels } = useLabels({ enabled: labelsShown });
  const labelNames = useMemo(
    () => new Map((labels ?? []).map((label) => [String(label.id), label.name])),
    [labels]
  );
  const statuses = statusesQuery.data;
  const sortRows = useCallback(
    (tasks: Task[]) => sortListTasks(tasks, prefs.sort, { statuses: statuses ?? [], labelNames }),
    [prefs.sort, statuses, labelNames]
  );
  // Sorted first, then grouped: grouping keeps the order inside each group.
  const groups = useMemo(
    () => buildListGroups(sortRows(shownTasks), statuses ?? [], groupBy, t),
    [sortRows, shownTasks, statuses, groupBy, t]
  );

  // Open parents, for this visit. Kept here rather than in the rows, which
  // unmount as they scroll out of a long list.
  const [expanded, setExpanded] = useState<ReadonlySet<string>>(() => new Set());
  useEffect(() => setExpanded(new Set()), [projectId]);
  const toggleExpanded = useCallback((taskId: string) => {
    setExpanded((current) => {
      const next = new Set(current);
      if (next.has(taskId)) next.delete(taskId);
      else next.add(taskId);
      return next;
    });
  }, []);
  const grid = useMemo<ListGridState>(
    () => ({ columns: prefs.columns, widths: prefs.widths, expanded, toggleExpanded, sortRows }),
    [prefs.columns, prefs.widths, expanded, toggleExpanded, sortRows]
  );
  const gridColumn = useRef(0);

  const sortBy = (column: ListColumnId): void =>
    setPrefs((current) => ({ ...current, sort: nextSort(current.sort, column) }));
  const resize = (column: ListColumnId, width: number): void =>
    setPrefs((current) =>
      current.widths[column] === width
        ? current
        : { ...current, widths: { ...current.widths, [column]: width } }
    );

  // What the export holds: the rows on screen in their order, less any that
  // only show faded because "Dim instead of hide" is on.
  const exportRows = useMemo(
    () => groups.flatMap((group) => group.tasks).filter((task) => !dimmedIds?.has(task.id)),
    [groups, dimmedIds]
  );
  const exportCsv = (): void => {
    const csv = listTasksToCsv(exportRows, prefs.columns, {
      t,
      statuses: statuses ?? [],
      labelNames,
    });
    downloadCsv(csvFileName(project?.name ?? '', todayIso()), csv);
  };

  if (
    (projectLoading && !project) ||
    (tasksQuery.isPending && !hasQueryFailed(tasksQuery)) ||
    (statusesQuery.isPending && !hasQueryFailed(statusesQuery))
  ) {
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

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectList.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={projectTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          <>
            <ListColumnPicker prefs={prefs} onChange={setPrefs} />
            <ExportCsvButton disabled={exportRows.length === 0} onExport={exportCsv} />
            <div className="w-[180px]">
              <FlowSelect
                label={t('flow.projectList.groupBy')}
                value={groupBy}
                options={LIST_GROUP_KEYS.map((key) => ({
                  value: key,
                  label: t(`flow.projectList.groupByOption.${key}`),
                }))}
                onChange={(value) => setGroupBy(value as ListGroupKey)}
              />
            </div>
          </>
        }
      />

      <FilterBar
        state={projectFilters.state}
        currentUserId={projectFilters.currentUserId}
        fields={projectFilters.fields}
        people={projectFilters.people}
        visible={matchingIds?.size ?? projectTasks.length}
        total={projectTasks.length}
        savingView={saveAsView.pending}
        onSaveAsView={() =>
          saveAsView.save(
            t('flow.filters.savedViewName', { project: project.name }),
            projectFiltersForView(projectFilters.state.filters, projectId, statusesQuery.data ?? [])
          )
        }
      />

      <div className="flex-1 overflow-auto">
        <ListGridContext.Provider value={grid}>
          <div
            role="treegrid"
            aria-label={t('flow.projectList.gridLabel', { project: project.name })}
            aria-colcount={prefs.columns.length + 2}
            style={{ minWidth: rowMinWidth(prefs) }}
            onKeyDown={(event) => handleGridKeyDown(event, gridColumn)}
          >
            <ListHeader prefs={prefs} onSort={sortBy} onResize={resize} />
            {groups.map((group) => (
              <TaskGroup
                key={`${groupBy}:${group.id}`}
                group={group}
                showHeader={groupBy !== 'none'}
                dimmedIds={dimmedIds}
              />
            ))}
          </div>
        </ListGridContext.Provider>
        {groups.length === 0 && (
          <p className="px-4 py-6 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {projectTasks.length > 0 && matchingIds !== null
              ? t('flow.filters.noMatches')
              : t('flow.projectList.empty')}
          </p>
        )}
      </div>
    </div>
  );
};

export default ProjectList;
