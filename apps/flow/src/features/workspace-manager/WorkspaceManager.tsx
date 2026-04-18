import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Users, Plus, Trash2, Shield, Eye, Crown } from 'lucide-react';
import { toast } from 'sonner';
import { useParams } from 'react-router-dom';
import {
  useWorkspaceMembers,
  useAddWorkspaceMember,
  useRemoveWorkspaceMember,
  useUpdateWorkspaceMemberRole,
} from './useWorkspaceManager';
import type { ApiWorkspaceMember } from '../../shared/services/flowApi';

const ROLE_LABELS: Record<string, { label: string; icon: React.ReactNode; color: string }> = {
  manager: { label: 'Manager', icon: <Crown size={11} />, color: '#f59e0b' },
  member: { label: 'Member', icon: <Shield size={11} />, color: '#64C5B9' },
  viewer: { label: 'Viewer', icon: <Eye size={11} />, color: '#8b5cf6' },
};

const MemberRow: React.FC<{
  member: ApiWorkspaceMember;
  workspaceId: string;
  isLastManager: boolean;
}> = ({ member, workspaceId, isLastManager }) => {
  const remove = useRemoveWorkspaceMember(workspaceId);
  const updateRole = useUpdateWorkspaceMemberRole(workspaceId);
  const roleInfo = ROLE_LABELS[member.role] ?? ROLE_LABELS.member;
  const initials = member.name
    ? member.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : member.email.slice(0, 2).toUpperCase();

  return (
    <div
      className="flex items-center gap-3 px-4 py-3 rounded-lg"
      style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
    >
      <span
        className="h-8 w-8 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0"
        style={{ background: 'rgba(100,197,185,0.15)', color: '#64C5B9' }}
      >
        {initials}
      </span>
      <div className="flex-1 min-w-0">
        <p className="text-[12px] font-medium truncate" style={{ color: 'var(--flow-text)' }}>
          {member.name || member.email}
        </p>
        {member.name && (
          <p className="text-[10px] truncate" style={{ color: 'var(--flow-text-dim)' }}>
            {member.email}
          </p>
        )}
      </div>
      <div className="flex items-center gap-1 shrink-0" style={{ color: roleInfo.color }}>
        {roleInfo.icon}
        <select
          value={member.role}
          onChange={(e) => {
            updateRole.mutate(
              {
                workspace_id: Number(workspaceId),
                user_id: Number(member.userId),
                role: e.target.value,
              },
              {
                onSuccess: () => toast.success('Role updated'),
                onError: () => toast.error('Failed to update role'),
              }
            );
          }}
          className="text-[11px] outline-none bg-transparent cursor-pointer"
          style={{ color: roleInfo.color }}
        >
          <option value="manager">Manager</option>
          <option value="member">Member</option>
          <option value="viewer">Viewer</option>
        </select>
      </div>
      <button
        onClick={() => {
          if (isLastManager) return toast.error('Cannot remove the last manager');
          remove.mutate(
            { workspace_id: Number(workspaceId), user_id: Number(member.userId) },
            {
              onSuccess: () => toast.success('Member removed'),
              onError: () => toast.error('Failed to remove member'),
            }
          );
        }}
        className="h-7 w-7 rounded flex items-center justify-center shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
        style={{ color: 'var(--flow-text-dim)' }}
        onMouseEnter={(e) => (e.currentTarget.style.color = '#ef4444')}
        onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--flow-text-dim)')}
        title="Remove member"
        disabled={isLastManager}
      >
        <Trash2 size={11} />
      </button>
    </div>
  );
};

const WorkspaceManager: React.FC = () => {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const wsId = workspaceId ?? '';
  const { data: members = [], isLoading } = useWorkspaceMembers(wsId);
  const addMember = useAddWorkspaceMember(wsId);

  const [inviteUserId, setInviteUserId] = useState('');
  const [inviteRole, setInviteRole] = useState<'manager' | 'member' | 'viewer'>('member');
  const [showInvite, setShowInvite] = useState(false);

  const managerCount = members.filter((m) => m.role === 'manager').length;

  const handleInvite = () => {
    const userId = Number(inviteUserId.trim());
    if (!userId) return toast.error('Enter a valid user ID');
    addMember.mutate(
      { workspace_id: Number(wsId), user_id: userId, role: inviteRole },
      {
        onSuccess: () => {
          toast.success('Member added');
          setInviteUserId('');
          setShowInvite(false);
        },
        onError: () => toast.error('Failed to add member'),
      }
    );
  };

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users size={14} style={{ color: 'var(--flow-text-muted)' }} />
            <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
              Workspace Members
            </h1>
          </div>
          <button
            onClick={() => setShowInvite((v) => !v)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
            style={{ background: '#64C5B9', color: '#0a0a0a' }}
          >
            <Plus size={11} /> Add member
          </button>
        </div>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-dim)' }}>
          {members.length} member{members.length !== 1 ? 's' : ''} in this workspace
        </p>
      </div>

      <div className="p-6 max-w-2xl space-y-3">
        {showInvite && (
          <motion.div
            initial={{ opacity: 0, y: -6 }}
            animate={{ opacity: 1, y: 0 }}
            className="rounded-lg p-4 space-y-3"
            style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
          >
            <p className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
              Add member by user ID
            </p>
            <div className="flex gap-2">
              <input
                type="number"
                placeholder="User ID"
                value={inviteUserId}
                onChange={(e) => setInviteUserId(e.target.value)}
                className="flex-1 rounded px-3 py-2 text-[12px] outline-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              />
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value as 'manager' | 'member' | 'viewer')}
                className="rounded px-3 py-2 text-[12px] outline-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              >
                <option value="manager">Manager</option>
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
              </select>
              <button
                onClick={handleInvite}
                disabled={addMember.isPending}
                className="px-4 py-2 rounded text-[12px] font-semibold"
                style={{ background: '#64C5B9', color: '#0a0a0a' }}
              >
                {addMember.isPending ? 'Adding...' : 'Add'}
              </button>
            </div>
          </motion.div>
        )}

        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-14 rounded-lg animate-pulse"
                style={{ background: 'var(--flow-hover)' }}
              />
            ))}
          </div>
        ) : members.length === 0 ? (
          <div className="text-center py-12 text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
            No members yet. Add someone to get started.
          </div>
        ) : (
          <div className="space-y-2 group">
            {members.map((m) => (
              <MemberRow
                key={m.id}
                member={m}
                workspaceId={wsId}
                isLastManager={m.role === 'manager' && managerCount <= 1}
              />
            ))}
          </div>
        )}

        <div
          className="rounded-lg p-4 mt-4"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <p
            className="text-[11px] font-semibold uppercase tracking-wider mb-2"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Roles
          </p>
          <div className="space-y-1.5 text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            <div className="flex items-center gap-2">
              <Crown size={11} style={{ color: '#f59e0b' }} />
              <strong style={{ color: 'var(--flow-text)' }}>Manager</strong>, can create/archive
              projects, manage all members
            </div>
            <div className="flex items-center gap-2">
              <Shield size={11} style={{ color: '#64C5B9' }} />
              <strong style={{ color: 'var(--flow-text)' }}>Member</strong>, can create/edit tasks
              in their projects
            </div>
            <div className="flex items-center gap-2">
              <Eye size={11} style={{ color: '#8b5cf6' }} />
              <strong style={{ color: 'var(--flow-text)' }}>Viewer</strong>, read-only access
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkspaceManager;
