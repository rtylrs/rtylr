import React, { useMemo, useState, useEffect } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import {
  Plus,
  Calendar,
  ChevronRight,
  Check,
  LayoutGrid,
  List,
  MoreHorizontal,
  ChevronDown,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore, Task } from '../../shared/stores/flowStore';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { cn, formatDueDate } from '../../shared/lib/utils';
import { useTasks, useUpdateTaskStatus } from '../tasks/useTasks';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCurrentProject } from './useCurrentProject';
import { useStatusLabels } from '../../shared/hooks/useStatusLabels';
import { TaskContextMenu, ContextMenuState } from '../../shared/components/TaskContextMenu';
import PriorityIcon from '../../shared/components/PriorityIcon';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

interface StatusGroupDef {
  id: Task['status'];
  label: string;
  color: string;
  alwaysShown: boolean;
}

// Labels are filled in from useStatusLabels() at render time (the single
// i18n-backed source of truth - see constantsFlow.ts's STATUS_LABEL_KEYS) so
// this view can never hardcode a different string than the rest of the app.
// Every value of the status enum has a group. backlog and cancelled were
// absent, so a task in either state rendered in no group at all and was simply
// invisible on this screen while showing up on the board (#1049).
//
// `alwaysShown` keeps the appearance of a project that has neither: those two
// groups appear only when they hold something, so a list that looked like four
// groups yesterday still does.
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
        onClick={handleToggle}
        aria-label={t('flow.projectList.toggleDoneLabel')}
        aria-pressed={checked}
        className="h-4 w-4 rounded-sm shrink-0 flex items-center justify-center transition-colors border"
        style={{
          background: checked ? 'var(--accent)' : 'transparent',
          borderColor: checked ? 'var(--accent)' : 'var(--flow-border)',
        }}
      >
        {checked && <Check size={10} strokeWidth={3} color="var(--flow-on-accent)" />}
      </button>

      {/* The same glyph the board uses for the same task. This was a 6px
          filled circle, identical in shape and size at every priority, so the
          only signal was the tint: WCAG 1.4.1, and the two views of one
          project encoded priority differently. PriorityIcon carries the shape,
          the tint and an aria-label. */}
      <span className="shrink-0">
        <PriorityIcon priority={task.priority} size={12} />
      </span>

      {/* Title */}
      <span
        className={cn('flex-1 text-body font-medium truncate', checked ? 'line-through' : '')}
        style={{ color: checked ? 'var(--flow-text-dim)' : 'var(--flow-text)' }}
      >
        {task.title}
      </span>

      {/* Subtask indicator */}
      {task.subtaskCount > 0 && (
        <span
          className="text-meta flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <ChevronRight size={11} />
          {task.subtaskCount} subtasks
        </span>
      )}

      {/* Assignee */}
      <span
        className="h-5 w-5 rounded-full flex items-center justify-center text-meta font-bold shrink-0"
        style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
      >
        {task.assigneeInitials}
      </span>

      {/* Due date */}
      <span
        className="text-meta flex items-center gap-1 w-[96px] justify-end"
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
  /**
   * Whether to show the task count beside the status name.
   *
   * Off when the list is split into sections. There the count is that
   * section's, while the board's identically labelled count is the whole
   * project's, so the two read as a contradiction with nothing on screen
   * saying which is which: "To Do 28" against "To Do 49" for the same status
   * (#1166). Both are right. The section total one line above already gives
   * the number, so the group count is what goes.
   *
   * On for an unsectioned list, where the group count and the board's count
   * are the same number and there is nothing to confuse.
   */
  showCount?: boolean;
}> = ({ group, tasks, showCount = true }) => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);

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
        <span className="text-meta font-semibold" style={{ color: group.color }}>
          {group.label}
        </span>
        {showCount && (
          <span className="text-meta ml-1 tabular-nums" style={{ color: 'var(--flow-text-dim)' }}>
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

            {/* New task row */}
            <div
              className="flex items-center gap-3 px-4 py-2 hover:bg-[var(--flow-hover)] transition-colors cursor-pointer opacity-50 hover:opacity-100"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
              onClick={() => useFlowStore.getState().setQuickCreateOpen(true)}
            >
              <div
                className="h-4 w-4 rounded-sm border flex items-center justify-center"
                style={{ borderColor: 'var(--flow-border)' }}
              >
                <Plus size={10} style={{ color: 'var(--flow-text-dim)' }} />
              </div>
              <span className="text-body" style={{ color: 'var(--flow-text-dim)' }}>
                {t('flow.shell.newTask')}
              </span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

const ProjectList: React.FC = () => {
  const { t } = useTranslation();
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = 'p-web' } = useParams({ strict: false });
  const statusLabels = useStatusLabels();
  const STATUS_GROUPS: StatusGroupDef[] = useMemo(
    () => STATUS_GROUP_COLORS.map((def) => ({ ...def, label: statusLabels[def.id] })),
    [statusLabels]
  );

  useProjectRealtime(projectId);

  const { tasks, sections } = useFlowStore();
  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectLoadFailed,
    refetch: refetchProjects,
  } = useCurrentProject(workspaceId, projectId);
  const { data: apiTasks } = useTasks({ projectId, activeOnly: true, compact: true });

  // `activeOnly` is a query param the backend doesn't currently act on, so
  // archived tasks still come back over the wire - filter client-side to
  // match Triage/FocusToday/FlowLayout.
  const projectTasks: Task[] = (
    apiTasks ? apiTasks.map(apiTaskToTask) : tasks.filter((t) => t.projectId === projectId)
  ).filter((t) => !t.archived);

  const projectSections = useMemo(
    () => sections.filter((s) => s.projectId === projectId).sort((a, b) => a.order - b.order),
    [sections, projectId]
  );

  const sectionBlocks = useMemo(() => {
    if (projectSections.length === 0) return null;
    const blocks: { id: string; title: string; tasks: Task[] }[] = projectSections.map((sec) => ({
      id: sec.id,
      title: sec.name,
      tasks: projectTasks.filter((task) => task.sectionId === sec.id),
    }));
    const inSection = new Set(projectSections.map((s) => s.id));
    const loose = projectTasks.filter((task) => !task.sectionId || !inSection.has(task.sectionId));
    if (loose.length)
      blocks.push({ id: '_unsectioned', title: t('flow.projectList.otherSection'), tasks: loose });
    return blocks;
  }, [projectSections, projectTasks, t]);

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

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectList.pageTitle')} />
      {/* Header */}
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div className="flex items-center gap-3">
          <span className="h-3 w-3 rounded-full" style={{ background: project.color }} />
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {project.name}
          </h1>
        </div>

        {/* View Switcher */}
        <div
          className="flex items-center rounded-lg p-0.5"
          style={{
            background: 'var(--flow-surface)',
            border: '1px solid var(--flow-border)',
          }}
        >
          <Link
            to="/p/$projectId"
            params={{ projectId }}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-meta font-medium transition-colors hover:bg-[var(--flow-hover)] no-underline"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <LayoutGrid size={13} />
            {t('flow.projectViewTabs.board')}
          </Link>
          <button
            type="button"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-meta font-medium transition-colors"
            style={{
              background: 'var(--accent-muted)',
              color: 'var(--accent-on-muted)',
            }}
          >
            <List size={13} />
            {t('flow.projectViewTabs.list')}
          </button>
        </div>
      </div>

      {/* Column headers */}
      <div
        className="flex items-center gap-3 px-4 py-2 text-meta font-semibold uppercase tracking-wider"
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
        <div className="w-[96px] text-right">{t('common.dueDate')}</div>
        <div className="w-5" />
      </div>

      {/* Sections + status groups */}
      <div className="flex-1 overflow-y-auto">
        {sectionBlocks
          ? sectionBlocks.map((block) => (
              <div key={block.id} className="mb-4">
                <div
                  className="px-4 py-2 text-meta font-semibold uppercase tracking-wider sticky top-0 z-[1]"
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
                {STATUS_GROUPS.map((group) => {
                  const groupTasks = block.tasks.filter((t) => t.status === group.id);
                  if (!groupIsVisible(group, groupTasks.length)) return null;
                  return (
                    <StatusGroup
                      key={`${block.id}-${group.id}`}
                      group={group}
                      tasks={groupTasks}
                      showCount={false}
                    />
                  );
                })}
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
