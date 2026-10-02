import React from 'react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle, ArrowUp, Equal, ArrowDown, Minus } from 'lucide-react';
import { Task, PRIORITY_COLORS } from '../stores/flowStore';
import { priorityLabelKey } from '../lib/priorityLabels';

interface PriorityIconProps {
  priority: Task['priority'];
  size?: number;
  showLabel?: boolean;
}

const ICONS: Record<Task['priority'], React.ComponentType<{ size?: number }>> = {
  urgent: AlertTriangle,
  high: ArrowUp,
  medium: Equal,
  low: ArrowDown,
  no_priority: Minus,
};

export const PriorityIcon: React.FC<PriorityIconProps> = ({
  priority,
  size = 13,
  showLabel = false,
}) => {
  const { t } = useTranslation();
  const Icon = ICONS[priority] ?? Minus;
  const color = PRIORITY_COLORS[priority] ?? 'var(--flow-text-dim)';
  const label = t(priorityLabelKey(priority));
  return (
    <span
      className="inline-flex items-center gap-1.5"
      style={{ color }}
      title={label}
      // role and aria-label, not title alone: a title is announced
      // inconsistently and never on touch, and WCAG 1.4.1 needs the priority
      // available without relying on the tint.
      role={showLabel ? undefined : 'img'}
      aria-label={showLabel ? undefined : label}
    >
      <Icon size={size} aria-hidden />
      {showLabel && <span className="text-flow-meta font-medium">{label}</span>}
    </span>
  );
};

export default PriorityIcon;
