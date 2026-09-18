import React from 'react';
import { useTranslation } from 'react-i18next';
import { Circle, CheckCircle2, XCircle, Timer, CircleDot, Archive } from 'lucide-react';
import { Task, STATUS_COLORS } from '../stores/flowStore';
import { STATUS_LABEL_KEYS } from '../lib/constantsFlow';

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
  const { t } = useTranslation();
  const Icon = ICONS[status] ?? Circle;
  const color = STATUS_COLORS[status] ?? 'var(--flow-text-dim)';
  const label = STATUS_LABEL_KEYS[status] ? t(STATUS_LABEL_KEYS[status]) : status;
  return (
    <span className="inline-flex items-center gap-1.5" style={{ color }} title={label}>
      <Icon size={size} />
      {showLabel && <span className="text-meta font-medium">{label}</span>}
    </span>
  );
};

export default StatusIcon;
