import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Calendar, MessageSquare, Paperclip, GitBranch } from 'lucide-react';
import { Task, useFlowStore } from '../stores/flowStore';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import AssigneeAvatar from './AssigneeAvatar';
import LabelChip from './LabelChip';
import { TaskContextMenu, ContextMenuState } from './TaskContextMenu';

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
  const navigate = useNavigate();
  const { labels, comments, attachments, subtasks, projects } = useFlowStore();
  const taskLabels = labels.filter((l) => task.labelIds?.includes(l.id));
  const project = projects.find((p) => p.id === task.projectId);
  const commentCount = comments.filter((c) => c.taskId === task.id).length;
  const attachmentCount = attachments.filter((a) => a.taskId === task.id).length;
  const taskSubtasks = subtasks.filter((s) => s.taskId === task.id);
  const doneSubtasks = taskSubtasks.filter((s) => s.completed).length;
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const menuW = 160;
    const menuH = 120;
    const x = e.clientX + menuW > vw ? e.clientX - menuW : e.clientX;
    const y = e.clientY + menuH > vh ? e.clientY - menuH : e.clientY;
    setContextMenu({ x, y, taskId: task.id, projectId: task.projectId });
  };

  return (
    <>
      <motion.div
        layout
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.18 }}
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey) {
            onSelect?.();
          } else {
            navigate({ to: `/task/${task.id}` as string });
          }
        }}
        onContextMenu={handleContextMenu}
        className={cn('rounded-lg cursor-pointer group relative', compact ? 'p-2.5' : 'p-3.5')}
        style={{
          opacity: dimmed ? 0.5 : undefined,
          background: 'var(--flow-card)',
          border: selected ? '1px solid #64C5B9' : '1px solid var(--flow-border)',
          boxShadow: selected ? '0 0 0 3px var(--accent-muted)' : '0 1px 3px rgba(0,0,0,0.15)',
        }}
        whileHover={{ y: -1 }}
      >
        {/* Header row */}
        <div className="flex items-start justify-between gap-2 mb-2">
          <div className="flex min-w-0 items-center gap-2">
            <PriorityIcon priority={task.priority} />
            <span
              className="text-[10px] font-mono shrink-0"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              {task.identifier}
            </span>
            {task.statusName && (
              <span
                className="min-w-0 max-w-[7rem] shrink-0 truncate rounded-sm px-1.5 py-0.5 text-[10px] font-semibold"
                style={{
                  background: `${task.statusColor ?? '#64C5B9'}22`,
                  color: task.statusColor ?? 'var(--accent)',
                }}
                title={task.statusName}
              >
                {task.statusName}
              </span>
            )}
          </div>
          <div className="flex shrink-0 items-center gap-0.5">{controls}</div>
          {showProject && (project || task.projectName) && (
            <span
              className="min-w-0 max-w-[8rem] truncate rounded-sm px-1.5 py-0.5 text-[10px] font-semibold"
              style={{
                background: `${project?.color ?? '#64C5B9'}22`,
                color: project?.color ?? 'var(--accent)',
              }}
              title={project?.name ?? task.projectName}
            >
              {project?.name ?? task.projectName}
            </span>
          )}
        </div>

        {/* Title */}
        <p
          className={cn('font-medium leading-snug mb-2', compact ? 'text-[12px]' : 'text-[13px]')}
          style={{ color: 'var(--flow-text)' }}
        >
          {task.title}
        </p>

        {/* Labels */}
        {taskLabels.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {taskLabels.slice(0, 3).map((l) => (
              <LabelChip key={l.id} label={l} small />
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between mt-1">
          <div className="flex items-center gap-2">
            <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={18} />
            {taskSubtasks.length > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <GitBranch size={10} />
                {doneSubtasks}/{taskSubtasks.length}
              </span>
            )}
            {commentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <MessageSquare size={10} />
                {commentCount}
              </span>
            )}
            {attachmentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <Paperclip size={10} />
                {attachmentCount}
              </span>
            )}
          </div>

          {task.dueDate && (
            <span
              className="flex items-center gap-1 text-[10px]"
              style={{ color: overdue ? '#ef4444' : 'var(--flow-text-dim)' }}
            >
              <Calendar size={10} />
              {formatDate(task.dueDate)}
            </span>
          )}
        </div>
      </motion.div>

      {contextMenu && <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />}
    </>
  );
};

export default TaskCard;
