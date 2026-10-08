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
import { useTasks, useUpdateTaskStatus } from '../tasks/useTasks';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCurrentProject } from './useCurrentProject';
import { useProjectColumns } from './useColumns';
import { useStatusLabels } from '../../shared/hooks/useStatusLabels';
import { TaskContextMenu, ContextMenuState } from '../../shared/components/TaskContextMenu';
import PriorityIcon from '../../shared/components/PriorityIcon';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useCanEditTask, useCanCreateTask } from '../../shared/lib/currentUser';
import ProjectViewTabs from './ProjectViewTabs';

interface StatusGroupDef {
  id: Task['status'];
  label: string;
  color: string;
  alwaysShown: boolean;
}

export const STATUS_GROUP_COLORS: { id: Task['status']; color: string; alwaysShown: boolean }[] = [
  { id: 'backlog', color: 'var(--flow-neutral-fg)', alwaysShown: false },
  { id: 'todo', color: 'var(--flow-neutral-fg)', alwaysShown: true },
  { id: 'in_progress', color: 'var(--flow-info-fg)', alwaysShown: true },
  { id: 'in_review', color: 'var(--flow-warn-fg)', alwaysShown: true },
  { id: 'done', color: 'var(--flow-status-approved)', alwaysShown: true },
  { id: 'cancelled', color: 'var(--flow-neutral-fg)', alwaysShown: false },
];

/** Whether a group earns a header: one of the four the list always shows, or non-empty. */
export function groupIsVisible(group: { alwaysShown: boolean }, taskCount: number): boolean {
  return group.alwaysShown || taskCount > 0;
}

interface TaskRowProps {
  task: Task;
  depth?: number;
}

export const TaskRow: React.FC<TaskRowProps> = ({ task, depth = 0 }) => {
  const { t } = useTranslation();
  const [checked, setChecked] = useState(task.status === 'done');
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const updateTaskStatus = useUpdateTaskStatus();
  const openTask = useOpenTask();
  const canEdit = useCanEditTask(task.projectId);

  useEffect(() => {
    setChecked(task.status === 'done');
  }, [task.status]);

  const handleToggle = () => {
    const newChecked = !checked;
    setChecked(newChecked);
    const newStatus = newChecked ? 'done' : 'todo';
    updateTaskStatus.mutate(
      { id: task.id, status: newStatus },
      {
        onError: () => setChecked(!newChecked),
      }
    );
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
      layout
      initial={{ opacity: 0, x: -6 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.15 }}
      className="flex items-center gap-3 px-4 py-2.5 group hover:bg-[var(--flow-hover)] transition-colors cursor-pointer"
      style={{
        paddingLeft: depth > 0 ? `${16 + depth * 20}px` : undefined,
        borderBottom: '1px solid var(--flow-border)',
      }}
    >
      {/* Checkbox */}
      <button
        type="button"
        disabled={!canEdit}
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
          <ChevronRight size={11} />
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

export const StatusGroup: React.FC<{
  group: StatusGroupDef;
  tasks: Task[];
  showCount?: boolean;
}> = ({ group, tasks, showCount = true }) => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);
  const { projectId } = useParams({ strict: false });
  const canCreate = useCanCreateTask(projectId);

  return (
    <div className="mb-2">
      {/* Group header */}
      <button
        type="button"
        onClick={() => setCollapsed((c) => !c)}
        className="w-full flex items-center gap-2.5 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors"
      >
        <ChevronDown
          size={13}
          className={cn('transition-transform', collapsed ? '-rotate-90' : '')}
          style={{ color: 'var(--flow-text-dim)' }}
        />
        <span className="h-2 w-2 rounded-full" style={{ background: group.color }} />
        <span className="text-flow-meta font-semibold" style={{ color: group.color }}>
          {group.label}
        </span>
        {showCount && (
          <span
            className="text-flow-meta ml-1 tabular-nums"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {tasks.length}
          </span>
        )}
      </button>

      <AnimatePresence>
        {!collapsed && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.18 }}
            style={{ overflow: 'hidden' }}
          >
            {tasks.map((task) => (
              <TaskRow key={task.id} task={task} />
            ))}

            {canCreate && (
              <button
                type="button"
                className="w-full flex items-center gap-3 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors text-left"
                style={{
                  borderBottom: '1px solid var(--flow-border)',
                  color: 'var(--flow-text-muted)',
                }}
                onClick={() => useFlowStore.getState().setQuickCreateOpen(true)}
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
  const statusLabels = useStatusLabels();
  const canCreate = useCanCreateTask(projectId);
  const STATUS_GROUPS: StatusGroupDef[] = useMemo(
    () => STATUS_GROUP_COLORS.map((def) => ({ ...def, label: statusLabels[def.id] })),
    [statusLabels]
  );

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectMetadataFailed,
    refetch: refetchProjectMetadata,
  } = useCurrentProject(projectId);
  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: apiTasks } = tasksQuery;
  const columnsQuery = useProjectColumns(projectId);
  const projectLoadFailed =
    projectMetadataFailed || hasQueryFailed(tasksQuery) || hasQueryFailed(columnsQuery);
  const refetchProjects = () => {
    void refetchProjectMetadata();
    void tasksQuery.refetch();
    void columnsQuery.refetch();
  };

  const projectTasks = (apiTasks ?? []).map(apiTaskToTask).filter((task) => !task.archived);
  const projectSections = useMemo(
    () => [...(columnsQuery.data ?? [])].sort((a, b) => a.sortOrder - b.sortOrder),
    [columnsQuery.data]
  );

  const sectionBlocks = useMemo(() => {
    if (projectSections.length === 0) return null;
    // `Task.sectionId` is a stringified column id (see apiTaskToTask), so both
    // sides of these comparisons must be coerced. Comparing the raw number
    // matched `"5"` against `5` - always false - so every task fell into
    // "Other" while its own column rendered empty. The block id is
    // stringified too, because quick-create takes the section id as a string.
    const blocks: { id: string; title: string; tasks: Task[] }[] = projectSections.map((sec) => ({
      id: String(sec.id),
      title: sec.name,
      tasks: projectTasks.filter((task) => task.sectionId === String(sec.id)),
    }));
    const inSection = new Set(projectSections.map((s) => String(s.id)));
    const loose = projectTasks.filter((task) => !task.sectionId || !inSection.has(task.sectionId));
    if (loose.length)
      blocks.push({ id: '_unsectioned', title: t('flow.projectList.otherSection'), tasks: loose });
    return blocks;
  }, [projectSections, projectTasks, t]);

  if ((projectLoading && !project) || tasksQuery.isLoading || columnsQuery.isLoading) {
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
        <div className="w-[96px] text-right">{t('flowCommon.dueDate')}</div>
        <div className="w-5" />
      </div>

      {/* Sections + status groups */}
      <div className="flex-1 overflow-y-auto">
        {sectionBlocks
          ? sectionBlocks.map((block) => (
              <div key={block.id} className="mb-4">
                <div
                  className="px-4 py-2 text-flow-meta font-semibold uppercase tracking-wider sticky top-0 z-[1]"
                  style={{
                    background: 'var(--flow-bg)',
                    borderBottom: '1px solid var(--flow-border)',
                    color: 'var(--flow-text-muted)',
                  }}
                >
                  {block.title}
                  <span
                    className="ml-2 font-normal tabular-nums"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {block.tasks.length}
                  </span>
                </div>
                {block.tasks.map((task) => (
                  <TaskRow key={task.id} task={task} />
                ))}
                {canCreate && (
                  <button
                    type="button"
                    className="w-full flex items-center gap-3 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors text-left"
                    style={{ color: 'var(--flow-text-muted)' }}
                    onClick={() =>
                      useFlowStore
                        .getState()
                        .setQuickCreateOpen(
                          true,
                          block.id === '_unsectioned' ? undefined : block.id
                        )
                    }
                  >
                    <Plus size={14} />
                    <span className="text-flow-body">{t('flow.shell.newTask')}</span>
                  </button>
                )}
              </div>
            ))
          : STATUS_GROUPS.map((group) => {
              const groupTasks = projectTasks.filter((t) => t.status === group.id);
              if (!groupIsVisible(group, groupTasks.length)) return null;
              return <StatusGroup key={group.id} group={group} tasks={groupTasks} />;
            })}
      </div>
    </div>
  );
};

export default ProjectList;
