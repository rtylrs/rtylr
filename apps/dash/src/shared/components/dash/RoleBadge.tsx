import React from 'react';
import { cn } from '../../lib/utils';

export interface RoleBadgeProps {
  role: string;
  className?: string;
}

const LABELS: Record<string, string> = {
  owner: 'Owner',
  admin: 'Admin',
  operator: 'Operator',
  billing: 'Billing',
  menu_manager: 'Menu Manager',
  staff: 'Staff',
  viewer: 'Viewer',
};

const COLORS: Record<string, { bg: string; fg: string; border: string }> = {
  owner: {
    bg: 'var(--status-cancelled-bg)',
    fg: 'var(--status-cancelled)',
    border: 'var(--status-cancelled-bg)',
  },
  admin: { bg: 'var(--accent-muted)', fg: 'var(--accent)', border: 'var(--accent-border)' },
  operator: { bg: 'rgba(139,92,246,0.14)', fg: 'var(--app-crm)', border: 'rgba(139,92,246,0.3)' },
  billing: {
    bg: 'var(--status-trial-bg)',
    fg: 'var(--status-trial)',
    border: 'var(--status-trial-bg)',
  },
  menu_manager: {
    bg: 'rgba(100,197,185,0.14)',
    fg: 'var(--app-flow)',
    border: 'rgba(100,197,185,0.3)',
  },
  staff: {
    bg: 'var(--status-active-bg)',
    fg: 'var(--status-active)',
    border: 'var(--status-active-bg)',
  },
  viewer: { bg: 'var(--dash-hover)', fg: 'var(--dash-text-muted)', border: 'var(--dash-border)' },
};

const DEFAULT = {
  bg: 'var(--dash-hover)',
  fg: 'var(--dash-text-dim)',
  border: 'var(--dash-border)',
};

export const RoleBadge: React.FC<RoleBadgeProps> = ({ role, className }) => {
  const c = COLORS[role.toLowerCase()] ?? DEFAULT;
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider',
        className
      )}
      style={{ background: c.bg, color: c.fg, border: `1px solid ${c.border}` }}
    >
      {LABELS[role.toLowerCase()] ?? role}
    </span>
  );
};

export default RoleBadge;
