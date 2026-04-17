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
  owner: { bg: 'rgba(239,68,68,0.12)', fg: '#ef4444', border: 'rgba(239,68,68,0.25)' },
  admin: { bg: 'rgba(69,131,196,0.14)', fg: '#4583C4', border: 'rgba(69,131,196,0.3)' },
  operator: { bg: 'rgba(139,92,246,0.14)', fg: '#8b5cf6', border: 'rgba(139,92,246,0.3)' },
  billing: { bg: 'rgba(245,158,11,0.14)', fg: '#f59e0b', border: 'rgba(245,158,11,0.3)' },
  menu_manager: { bg: 'rgba(100,197,185,0.14)', fg: '#64C5B9', border: 'rgba(100,197,185,0.3)' },
  staff: { bg: 'rgba(16,185,129,0.14)', fg: '#10b981', border: 'rgba(16,185,129,0.3)' },
  viewer: { bg: 'rgba(156,163,175,0.18)', fg: '#6b7280', border: 'rgba(156,163,175,0.3)' },
};

const DEFAULT = { bg: 'rgba(100,197,185,0.14)', fg: '#64C5B9', border: 'rgba(100,197,185,0.3)' };

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
