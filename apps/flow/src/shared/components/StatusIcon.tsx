import React from 'react';
import { Circle, CheckCircle2, XCircle, Timer, CircleDot, Archive } from 'lucide-react';
import { Task, STATUS_COLORS, STATUS_LABELS } from '../stores/flowStore';

interface StatusIconProps {
  status: Task['status'];
  size?: number;
  showLabel?: boolean;
}

const ICONS: Record<Task['status'], React.ComponentType<{ size?: number }>> = {
  backlog: Archive,
  todo: Circle,
  in_progress: Timer,
  in_review: CircleDot,
  done: CheckCircle2,
  cancelled: XCircle,
};

export const StatusIcon: React.FC<StatusIconProps> = ({ status, size = 13, showLabel = false }) => {
  const Icon = ICONS[status] ?? Circle;
  const color = STATUS_COLORS[status] ?? 'var(--flow-text-dim)';
  const label = STATUS_LABELS[status] ?? status;
  return (
    <span className="inline-flex items-center gap-1.5" style={{ color }} title={label}>
      <Icon size={size} />
      {showLabel && <span className="text-[12px] font-medium">{label}</span>}
    </span>
  );
};

export default StatusIcon;
