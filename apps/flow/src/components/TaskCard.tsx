import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Calendar, MessageSquare, Paperclip, GitBranch } from 'lucide-react';
import { Task, useFlowStore } from '../stores/flowStore';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import AssigneeAvatar from './AssigneeAvatar';
import LabelChip from './LabelChip';

interface TaskCardProps {
  task: Task;
  compact?: boolean;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  selected?: boolean;
  onSelect?: () => void;
}

export const TaskCard: React.FC<TaskCardProps> = ({
  task,
  compact,
  draggable,
  onDragStart,
  selected,
  onSelect,
}) => {
  const navigate = useNavigate();
  const { labels, comments, attachments, subtasks } = useFlowStore();
  const taskLabels = labels.filter((l) => task.labels?.includes(l.id));
  const commentCount = comments.filter((c) => c.taskId === task.id).length;
  const attachmentCount = attachments.filter((a) => a.taskId === task.id).length;
  const taskSubtasks = subtasks.filter((s) => s.taskId === task.id);
  const doneSubtasks = taskSubtasks.filter((s) => s.completed).length;

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -8 }}
      transition={{ duration: 0.18 }}
      draggable={draggable}
      onDragStart={onDragStart}
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey) {
          onSelect?.();
        } else {
          navigate(`/task/${task.id}`);
        }
      }}
      className={cn(
        'rounded-lg cursor-pointer group relative',
        compact ? 'p-2.5' : 'p-3.5'
      )}
      style={{
        background: 'var(--flow-card)',
        border: selected
          ? '1px solid #64C5B9'
          : '1px solid var(--flow-border)',
        boxShadow: selected
          ? '0 0 0 3px rgba(100,197,185,0.15)'
          : '0 1px 3px rgba(0,0,0,0.15)',
      }}
      whileHover={{ y: -1 }}
    >
      {/* Header row */}
      <div className="flex items-start gap-2 mb-2">
        <PriorityIcon priority={task.priority} />
        <span
          className="text-[10px] font-mono shrink-0"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          {task.identifier}
        </span>
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
          <AssigneeAvatar
            name={task.assignee}
            initials={task.assigneeInitials}
            size={18}
          />
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
  );
};

export default TaskCard;
