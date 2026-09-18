import React from 'react';
import { Link } from '@tanstack/react-router';
import { Mail } from 'lucide-react';
import { StaffMember } from '../services/hrApi';

const PALETTE = [
  { bg: 'rgba(139,92,246,0.15)', fg: '#8b5cf6' },
  { bg: 'rgba(16,185,129,0.15)', fg: '#10b981' },
  { bg: 'rgba(59,130,246,0.15)', fg: '#3b82f6' },
  { bg: 'rgba(245,158,11,0.15)', fg: '#f59e0b' },
  { bg: 'rgba(239,68,68,0.15)', fg: '#ef4444' },
  { bg: 'rgba(100,197,185,0.15)', fg: '#64C5B9' },
];

export function initials(name: string) {
  return name
    .split(' ')
    .map((n) => n[0] ?? '')
    .join('')
    .slice(0, 2)
    .toUpperCase();
}

export function paletteFor(id: number) {
  return PALETTE[id % PALETTE.length];
}

interface Props {
  staff: StaffMember;
  compact?: boolean;
  linkTo?: string;
}

const EmployeeCard: React.FC<Props> = ({ staff, compact = false, linkTo }) => {
  const { bg, fg } = paletteFor(staff.id);
  const href = linkTo ?? `/staff/${staff.id}`;

  const body = (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: compact ? 10 : 14,
        borderRadius: 12,
        border: '1px solid var(--hr-border)',
        background: 'var(--hr-card)',
        transition: 'border-color 0.15s, background 0.15s',
        textDecoration: 'none',
        color: 'inherit',
      }}
      onMouseEnter={(e) =>
        ((e.currentTarget as HTMLDivElement).style.borderColor = 'rgba(139,92,246,0.35)')
      }
      onMouseLeave={(e) =>
        ((e.currentTarget as HTMLDivElement).style.borderColor = 'var(--hr-border)')
      }
    >
      <span
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: compact ? 32 : 40,
          height: compact ? 32 : 40,
          borderRadius: 10,
          background: bg,
          color: fg,
          fontSize: compact ? 11 : 13,
          fontWeight: 700,
          flexShrink: 0,
        }}
      >
        {initials(staff.name)}
      </span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 13,
            fontWeight: 600,
            color: 'var(--hr-text)',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
          }}
        >
          {staff.name}
        </div>
        <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
          {staff.role}
          {staff.department && <> • {staff.department}</>}
        </div>
      </div>
      {!compact && <Mail size={14} style={{ color: 'var(--hr-text-dim)', flexShrink: 0 }} />}
    </div>
  );

  return (
    <Link to={href} style={{ textDecoration: 'none' }}>
      {body}
    </Link>
  );
};

export default EmployeeCard;
