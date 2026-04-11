import React from 'react';
import {
  AlertTriangle,
  ArrowUp,
  Equal,
  ArrowDown,
  Minus,
} from 'lucide-react';
import { Task, PRIORITY_COLORS, PRIORITY_LABELS } from '../stores/flowStore';

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
  const Icon = ICONS[priority];
  const color = PRIORITY_COLORS[priority];
  return (
    <span
      className="inline-flex items-center gap-1.5"
      style={{ color }}
      title={PRIORITY_LABELS[priority]}
    >
      <Icon size={size} />
      {showLabel && (
        <span className="text-[12px] font-medium">{PRIORITY_LABELS[priority]}</span>
      )}
    </span>
  );
};

export default PriorityIcon;
