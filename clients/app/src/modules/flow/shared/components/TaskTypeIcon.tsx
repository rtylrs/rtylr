import React from 'react';
import { useTranslation } from 'react-i18next';
import { Bug, FlaskConical, Sparkles, Wrench, type LucideIcon } from 'lucide-react';
import type { TaskType } from '../stores/flowStore';

const ICONS: Partial<Record<TaskType, { icon: LucideIcon; color: string }>> = {
  bug: { icon: Bug, color: 'var(--flow-danger-fg)' },
  feature: { icon: Sparkles, color: 'var(--flow-status-approved)' },
  chore: { icon: Wrench, color: 'var(--flow-text-muted)' },
  spike: { icon: FlaskConical, color: 'var(--flow-info-fg)' },
};

/**
 * The task's type as an icon on cards and rows (T-380). Plain tasks show
 * nothing, and an epic already has its chip.
 */
const TaskTypeIcon: React.FC<{ type?: TaskType; size?: number; className?: string }> = ({
  type,
  size = 12,
  className,
}) => {
  const { t } = useTranslation();
  const entry = type ? ICONS[type] : undefined;
  if (!entry || !type) return null;
  const Icon = entry.icon;
  const label = t(`flow.taskType.${type}`);
  return (
    <span
      className={className ?? 'inline-flex shrink-0 items-center'}
      role="img"
      aria-label={label}
      title={label}
    >
      <Icon size={size} aria-hidden="true" style={{ color: entry.color }} />
    </span>
  );
};

export default TaskTypeIcon;
