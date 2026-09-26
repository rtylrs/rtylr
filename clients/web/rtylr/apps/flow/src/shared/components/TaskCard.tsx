import React, { useState } from 'react';
import { useOpenTask } from '../lib/useOpenTask';
import { dueState } from '../lib/dateUtils';
import { motion } from 'motion/react';
import { Calendar, MessageSquare, Paperclip, GitBranch } from 'lucide-react';
import { Task, useFlowStore } from '../stores/flowStore';
import { useLabels } from '../../features/tasks/useLabels';
import { useTranslation } from 'react-i18next';
import { labelsForTask } from '../lib/taskLabelUtils';
import { formatDueDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import AssigneeAvatar from './AssigneeAvatar';
import LabelChip from './LabelChip';
import { TaskContextMenu, ContextMenuState } from './TaskContextMenu';
import { withAlpha, onTint } from '../lib/colorUtils';

interface TaskCardProps {
  task: Task;
  compact?: boolean;
  showProject?: boolean;
  selected?: boolean;
  onSelect?: () => void;
  /**
   * Board-only controls (drag handle, move menu) rendered in the header.
   *
   * The card used to carry `draggable` and an `onDragStart` that had to be cast
   * through `unknown` to satisfy motion.div's own drag props. HTML5 drag never
   * fired on touch anyway, so the board owns the gesture now and passes its
   * affordances in here.
   */
  controls?: React.ReactNode;
  /** True while this card is the one being dragged, so it reads as lifted. */
  dimmed?: boolean;
}

export const TaskCard: React.FC<TaskCardProps> = ({
  task,
  compact,
  showProject,
  selected,
  onSelect,
  controls,
  dimmed,
}) => {
  const { t } = useTranslation();
  const openTask = useOpenTask();
  const { comments, attachments, subtasks, projects } = useFlowStore();
  // Labels come from the query, not the store. `setLabels` has no call site
  // anywhere in the app, so the store's `labels` was permanently [] and no
  // chip ever rendered here even when the task had labels.
  const { data: allLabels = [] } = useLabels();
  const taskLabels = labelsForTask(allLabels, task);
  const project = projects.find((p) => p.id === task.projectId);
  const commentCount = comments.filter((c) => c.taskId === task.id).length;
  const attachmentCount = attachments.filter((a) => a.taskId === task.id).length;
  const taskSubtasks = subtasks.filter((s) => s.taskId === task.id);
  const doneSubtasks = taskSubtasks.filter((s) => s.completed).length;
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const due = dueState(task.dueDate);

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
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
    <>
      <motion.div
        // `dimmed` goes through motion's own `animate`, not through `style`.
        //
        // This card used to set `style={{ opacity: dimmed ? 0.5 : undefined }}`
        // beside an animated `opacity`, and the two fought. React writes
        // `opacity: undefined` into the same style object motion is animating,
        // which pins the value at whatever `initial` set, so **every card on
        // the board stayed at opacity 0** and the columns, counts and empty
        // states rendered around invisible cards until some later re-render
        // shook them loose (rtylr#1078).
        //
        // Measured in headless and headed Chromium, on a dev server and a
        // production build: opacity 0 four seconds after load, rAF running,
        // `getAnimations()` empty. The inbox row has the same `layout` plus
        // opacity-0 `initial` and animates correctly, which is what ruled the
        // combination out as the cause and pointed at the competing style.
        layout
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: dimmed ? 0.5 : 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.18 }}
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey) {
            onSelect?.();
          } else {
            openTask(String(task.id));
          }
        }}
        onContextMenu={handleContextMenu}
        // p-3 rather than p-3.5: the rule is 10-12px and 14px was outside it.
        className={cn('rounded-lg cursor-pointer group relative', compact ? 'p-2.5' : 'p-3')}
        style={{
          background: 'var(--flow-card)',
          border: selected ? '1px solid var(--accent)' : '1px solid var(--flow-border)',
          boxShadow: selected ? '0 0 0 3px var(--accent-muted)' : 'var(--flow-shadow-card)',
        }}
        whileHover={{ y: -1 }}
      >
        {/* Header row */}
        <div className="flex items-start justify-between gap-2 mb-2">
          <div className="flex min-w-0 items-center gap-2">
            <PriorityIcon priority={task.priority} />
            <span
              className="text-meta font-mono shrink-0"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              {task.identifier}
            </span>
            {task.statusName && (
              <span
                className="min-w-0 max-w-[7rem] shrink-0 truncate rounded-sm px-1.5 py-0.5 text-meta font-semibold"
                style={{
                  // See Calendar: the tone is set once and read back, which also
                  // makes it observable in tests.
                  ...(task.statusColor ? { ['--tone' as string]: task.statusColor } : {}),
                  background: task.statusColor
                    ? withAlpha('var(--tone)', 13)
                    : 'var(--accent-muted)',
                  color: task.statusColor ? onTint('var(--tone)') : 'var(--accent-on-muted)',
                }}
                title={task.statusName}
              >
                {task.statusName}
              </span>
            )}
          </div>
          {/* Hover, and focus-within so the keyboard path survives. The drag
              handle is aria-hidden and pointer-only, but MoveToColumnMenu is
              how a keyboard or single-pointer user moves a card, so it has to
              come back on focus rather than only on hover (rtylr#1078). */}
          <div className="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100">
            {controls}
          </div>
          {showProject && (project || task.projectName) && (
            <span
              className="min-w-0 max-w-[8rem] truncate rounded-sm px-1.5 py-0.5 text-meta font-semibold"
              style={{
                // The optional chaining is why the rtylr#1246 sweep walked past
                // this one: the guard matched `color: project.color` and the
                // text here is `project?.color`.
                ...(project?.color ? { ['--tone' as string]: project.color } : {}),
                background: project?.color ? withAlpha('var(--tone)', 13) : 'var(--accent-muted)',
                color: project?.color ? onTint('var(--tone)') : 'var(--accent-on-muted)',
              }}
              title={project?.name ?? task.projectName}
            >
              {project?.name ?? task.projectName}
            </span>
          )}
        </div>

        {/* Title */}
        <p
          // Two lines, then an ellipsis. Unclamped, one long title made its card
          // several times the height of its neighbours and pushed the rest of
          // the column out of view (rtylr#1078, S-08). `title` carries the whole
          // string for anyone who needs it.
          className={cn(
            'font-medium leading-snug mb-2 line-clamp-2',
            compact ? 'text-meta' : 'text-body'
          )}
          style={{ color: 'var(--flow-text)' }}
          title={task.title}
        >
          {task.title}
        </p>

        {/* Labels */}
        {taskLabels.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {taskLabels.slice(0, 3).map((l) => (
              <LabelChip
                key={l.id}
                label={{ id: String(l.id), name: l.name, color: l.color }}
                small
              />
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between mt-1">
          <div className="flex items-center gap-2">
            <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={18} />
            {taskSubtasks.length > 0 && (
              <span
                className="flex items-center gap-0.5 text-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <GitBranch size={10} />
                {doneSubtasks}/{taskSubtasks.length}
              </span>
            )}
            {commentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <MessageSquare size={10} />
                {commentCount}
              </span>
            )}
            {attachmentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <Paperclip size={10} />
                {attachmentCount}
              </span>
            )}
          </div>

          {task.dueDate && (
            <span
              className="flex items-center gap-1 text-meta"
              // The same three states as TaskRow. This card kept --flow-text-dim
              // when the row was fixed, and that token measures 2.35:1 in both
              // modes on a declaration that says it never carries text.
              style={{
                color:
                  due === 'late'
                    ? 'var(--flow-danger-fg)'
                    : due === 'today'
                      ? 'var(--flow-warn-fg)'
                      : 'var(--flow-text-muted)',
              }}
            >
              <Calendar size={10} />
              {formatDueDate(task.dueDate, t)}
            </span>
          )}
        </div>
      </motion.div>

      {contextMenu && <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />}
    </>
  );
};

export default TaskCard;
