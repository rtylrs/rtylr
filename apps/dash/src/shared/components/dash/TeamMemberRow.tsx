import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { MoreHorizontal } from 'lucide-react';
import { Avatar } from '../ui';
import AppBadge from './AppBadge';
import RoleBadge from './RoleBadge';
import type { TeamMember } from '../../hooks/queries/useTeam';

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

function formatRelative(iso?: string): string {
  if (!iso) return 'never';
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export interface TeamMemberRowProps {
  member: TeamMember;
  onMenu?: (member: TeamMember) => void;
  index?: number;
}

export const TeamMemberRow: React.FC<TeamMemberRowProps> = ({ member, onMenu, index = 0 }) => {
  const navigate = useNavigate();

  return (
    <motion.div
      className="flex items-center gap-3 px-4 py-3 rounded-lg cursor-pointer hover:bg-[var(--dash-hover)]"
      onClick={() => navigate(`/team/${member.id}`)}
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1], delay: Math.min(index, 12) * 0.03 }}
    >
      <Avatar
        initials={getInitials(member.name)}
        size="md"
        className="shrink-0"
        style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
      />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="text-sm font-semibold truncate" style={{ color: 'var(--dash-text)' }}>
            {member.name}
          </p>
          <RoleBadge role={member.role} />
          {member.status !== 'active' && (
            <span
              className="inline-flex items-center rounded-full px-1.5 py-0.5 text-[9px] font-medium uppercase"
              style={{
                background:
                  member.status === 'invited' ? 'rgba(245,158,11,0.15)' : 'rgba(107,114,128,0.15)',
                color: member.status === 'invited' ? '#f59e0b' : '#6b7280',
              }}
            >
              {member.status}
            </span>
          )}
        </div>
        <p className="text-xs truncate" style={{ color: 'var(--dash-text-dim)' }}>
          {member.email} · active {formatRelative(member.lastActive)}
        </p>
      </div>
      <div className="hidden md:flex items-center gap-1.5 shrink-0">
        {member.apps.slice(0, 4).map((a) => (
          <AppBadge key={a} app={a} size="xs" showLabel={false} />
        ))}
        {member.apps.length > 4 && (
          <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
            +{member.apps.length - 4}
          </span>
        )}
      </div>
      <button
        className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-active)] shrink-0"
        style={{ color: 'var(--dash-text-dim)' }}
        onClick={(e) => {
          e.stopPropagation();
          onMenu?.(member);
        }}
        aria-label="Member actions"
      >
        <MoreHorizontal size={14} />
      </button>
    </motion.div>
  );
};

export default TeamMemberRow;
