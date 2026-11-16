import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import React, { useMemo, useState, useEffect } from 'react';
import { useParams } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { Plus, Calendar, ChevronRight, Check, MoreHorizontal, ChevronDown } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore, Task } from '../../shared/stores/flowStore';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { cn, formatDueDate } from '../../shared/lib/utils';
import { useTasks } from '../tasks/useTasks';
import { useToggleDone } from '../tasks/useToggleDone';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCurrentProject } from './useCurrentProject';
import { useProjectStatuses } from './useStatuses';
import { defaultStatus, sortStatuses, statusIdOf } from './statusUtils';
import { DUE_LANES, PRIORITY_LABEL_KEY, PRIORITY_ORDER, laneKeyFor } from './taskGroups';
import { TaskContextMenu, ContextMenuState } from '../../shared/components/TaskContextMenu';
import PriorityIcon from '../../shared/components/PriorityIcon';
import FlowSelect from '../../shared/components/FlowSelect';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useCanEditTask, useCanCreateTask } from '../../shared/lib/currentUser';
import ProjectViewTabs from './ProjectViewTabs';
import type { ApiStatus } from '../../shared/services/flowApi';
import type { RowProps } from '../../../../shared/hooks/useVirtualRows';
import { VirtualList } from '../../../../shared/components/VirtualList';

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
        label: s.name,
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

interface TaskRowProps {
  task: Task;
  depth?: number;
  /** From a windowed list (see VirtualList). */
  rowProps?: RowProps;
}

export const TaskRow: React.FC<TaskRowProps> = ({ task, depth = 0, rowProps }) => {
  // Windowed rows mount as they scroll into view, so they skip the entrance
  // and layout animations a short list uses.
  const windowed = rowProps?.['data-index'] !== undefined;
  const { t } = useTranslation();
  const [checked, setChecked] = useState(task.status === 'done');
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const toggleDone = useToggleDone(task);
  const openTask = useOpenTask();
  const canEdit = useCanEditTask(task.projectId);

  useEffect(() => {
    setChecked(task.status === 'done');
  }, [task.status]);

  const handleToggle = () => {
    const newChecked = !checked;
    setChecked(newChecked);
    const revert = () => setChecked(!newChecked);
    if (newChecked) toggleDone.complete(revert);
    else toggleDone.reopen(revert);
  };

  const handleMoreClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const menuW = 160;
    const menuH = 120;
    const x = e.clientX + menuW > vw ? e.clientX - menuW : e.clientX;
    const y = e.clientY + menuH > vh ? e.clientY - menuH : e.clientY;
    setContextMenu({ x, y, taskId: task.id, projectId: task.projectId, taskTitle: task.title });
  };

  return (
    <motion.div
      {...rowProps}
      role="button"
      tabIndex={0}
      aria-label={t('flowCommon.taskRow.openTask', {
        identifier: task.identifier,
        title: task.title,
      })}
      onClick={() => openTask(task.id)}
      onKeyDown={(event) => {
        if (event.target === event.currentTarget && (event.key === 'Enter' || event.key === ' ')) {
          event.preventDefault();
          openTask(task.id);
        }
      }}
      layout={!windowed}
      initial={windowed ? false : { opacity: 0, x: -6 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.15 }}
      className="flex items-center gap-3 px-4 py-2.5 group hover:bg-[var(--flow-hover)] transition-colors cursor-pointer"
      style={{
        paddingInlineStart: depth > 0 ? `${16 + depth * 20}px` : undefined,
        borderBottom: '1px solid var(--flow-border)',
      }}
    >
      {/* Checkbox */}
      <button
        type="button"
        disabled={!canEdit || !toggleDone.ready}
        onClick={(event) => {
          event.stopPropagation();
          handleToggle();
        }}
        aria-label={t('flow.projectList.toggleDoneLabel')}
        aria-pressed={checked}
        className="h-4 w-4 rounded-sm shrink-0 flex items-center justify-center transition-colors border"
        style={{
          background: checked ? 'var(--accent)' : 'transparent',
          borderColor: checked ? 'var(--accent)' : 'var(--flow-border)',
        }}
      >
        {checked && <Check size={10} strokeWidth={3} color="var(--dash-on-accent)" />}
      </button>

      <span className="shrink-0">
        <PriorityIcon priority={task.priority} size={12} />
      </span>

      {/* Title */}
      <span
        className={cn('flex-1 text-flow-body font-medium truncate', checked ? 'line-through' : '')}
        style={{ color: checked ? 'var(--flow-text-dim)' : 'var(--flow-text)' }}
      >
        {task.title}
      </span>

      {/* Subtask indicator */}
      {task.subtaskCount > 0 && (
        <span
          className="text-flow-meta flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <ChevronRight size={11} className="rtl:rotate-180" />
          {task.subtaskCount} {t('flow.taskDetail.subtasks')}
        </span>
      )}

      {/* Assignee */}
      <span
        className="h-5 w-5 rounded-full flex items-center justify-center text-flow-meta font-bold shrink-0"
        style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
      >
        {task.assigneeInitials}
      </span>

      {/* Due date */}
      <span
        className="text-flow-meta flex items-center gap-1 w-[96px] justify-end"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        <Calendar size={11} />
        {formatDueDate(task.dueDate, t)}
      </span>

      {/* More */}
      <button
        type="button"
        onClick={handleMoreClick}
        aria-label={t('flow.projectList.moreOptions')}
        className="opacity-0 group-hover:opacity-100 transition-opacity h-5 w-5 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        <MoreHorizontal size={13} />
      </button>

      {contextMenu && <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />}
    </motion.div>
  );
};

export const TaskGroup: React.FC<{
  group: ListGroup;
  /** False for the ungrouped list, which has no header. */
  showHeader?: boolean;
}> = ({ group, showHeader = true }) => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);
  const { projectId } = useParams({ strict: false });
  const canCreate = useCanCreateTask(projectId);
  const open = !showHeader || !collapsed;

  return (
    <div className="mb-2">
      {showHeader && (
        <button
          type="button"
          onClick={() => setCollapsed((c) => !c)}
          aria-expanded={!collapsed}
          className="w-full flex items-center gap-2.5 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors"
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
                />
              )}
            />

            {canCreate && (
              <button
                type="button"
                className="w-full flex items-center gap-3 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors text-start"
                style={{
                  borderBottom: '1px solid var(--flow-border)',
                  color: 'var(--flow-text-muted)',
                }}
                onClick={() => useFlowStore.getState().setQuickCreateOpen(true, group.statusId)}
              >
                <Plus size={14} />
                <span className="text-flow-body">{t('flow.shell.newTask')}</span>
              </button>
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
  const groups = useMemo(
    () => buildListGroups(projectTasks, statusesQuery.data ?? [], groupBy, t),
    [projectTasks, statusesQuery.data, groupBy, t]
  );

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
        }
      />

      {/* Column headers */}
      <div
        className="flex items-center gap-3 px-4 py-2 text-flow-meta font-semibold uppercase tracking-wider"
        style={{
          color: 'var(--flow-text-dim)',
          borderBottom: '1px solid var(--flow-border)',
          background: 'var(--flow-surface)',
        }}
      >
        <div className="w-4" />
        <div className="w-2" />
        <div className="flex-1">{t('flow.projectList.columnTitle')}</div>
        <div className="w-5" />
        <div className="w-[96px] text-end">{t('flowCommon.dueDate')}</div>
        <div className="w-5" />
      </div>

      <div className="flex-1 overflow-y-auto">
        {groups.map((group) => (
          <TaskGroup key={`${groupBy}:${group.id}`} group={group} showHeader={groupBy !== 'none'} />
        ))}
        {groups.length === 0 && (
          <p className="px-4 py-6 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.projectList.empty')}
          </p>
        )}
      </div>
    </div>
  );
};

export default ProjectList;
