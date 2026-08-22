import React, { useState, useMemo } from 'react';
import { Users, Plus, Trash2, Shield, Eye, Crown } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { toast } from 'sonner';
import { useParams } from '@tanstack/react-router';
import {
  useWorkspaceMembers,
  useRemoveWorkspaceMember,
  useUpdateWorkspaceMemberRole,
} from './useWorkspaceManager';
import AddWorkspaceMemberDialog from './AddWorkspaceMemberDialog';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import type { ApiWorkspaceMember } from '../../shared/services/flowApi';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { roleLabel } from '../../shared/lib/roleUtils';

// Icon and accent color per workspace role. Display text is not part of this
// map - it comes from roleLabel() + t(), the single source of truth for role
// names (see shared/lib/roleUtils.ts).
const ROLE_STYLE: Record<string, { icon: React.ReactNode; color: string }> = {
  admin: { icon: <Crown size={11} />, color: '#f59e0b' },
  member: { icon: <Shield size={11} />, color: 'var(--accent)' },
  viewer: { icon: <Eye size={11} />, color: '#8b5cf6' },
};

const WORKSPACE_ROLES: Array<'admin' | 'member' | 'viewer'> = ['admin', 'member', 'viewer'];

const MemberRow: React.FC<{
  member: ApiWorkspaceMember;
  workspaceId: string;
  isLastAdmin: boolean;
}> = ({ member, workspaceId, isLastAdmin }) => {
  const { t } = useTranslation();
  const remove = useRemoveWorkspaceMember(workspaceId);
  const updateRole = useUpdateWorkspaceMemberRole(workspaceId);
  const [confirmRemove, setConfirmRemove] = useState(false);
  const roleStyle = ROLE_STYLE[member.role] ?? ROLE_STYLE.member;
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
        style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
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
      <div className="flex items-center gap-1 shrink-0" style={{ color: roleStyle.color }}>
        {roleStyle.icon}
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
                onSuccess: () => toast.success(t('toasts.workspaceMember.roleUpdated')),
                onError: () => toast.error(t('toasts.workspaceMember.roleUpdateFailed')),
              }
            );
          }}
          className="text-[11px] outline-none bg-transparent cursor-pointer"
          style={{ color: roleStyle.color }}
        >
          {WORKSPACE_ROLES.map((role) => (
            <option key={role} value={role}>
              {t(roleLabel(role))}
            </option>
          ))}
        </select>
      </div>
      <button
        type="button"
        onClick={() => {
          if (isLastAdmin) return toast.error(t('toasts.workspaceMember.cannotRemoveLastAdmin'));
          setConfirmRemove(true);
        }}
        className="h-7 w-7 rounded-sm flex items-center justify-center shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
        style={{ color: 'var(--flow-text-dim)' }}
        onMouseEnter={(e) => (e.currentTarget.style.color = '#ef4444')}
        onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--flow-text-dim)')}
        title={t('workspaceManager.removeMemberTitle')}
        aria-label={t('workspaceManager.removeMemberTitle')}
        disabled={isLastAdmin}
      >
        <Trash2 size={11} />
      </button>
      <ConfirmDialog
        open={confirmRemove}
        onOpenChange={setConfirmRemove}
        title={t('workspaceManager.removeMemberTitle')}
        description={t('workspaceManager.removeMemberConfirm', {
          name: member.name || member.email,
        })}
        confirmLabel={t('workspaceManager.removeMemberAction')}
        cancelLabel={t('common.cancel')}
        pending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { workspace_id: Number(workspaceId), user_id: Number(member.userId) },
            {
              onSuccess: () => {
                toast.success(t('toasts.workspaceMember.removed'));
                setConfirmRemove(false);
              },
              onError: () => toast.error(t('toasts.workspaceMember.removeFailed')),
            }
          );
        }}
      />
    </div>
  );
};

const WorkspaceManager: React.FC = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ strict: false });
  const wsId = workspaceId ?? '';
  const { data: members = [], isLoading } = useWorkspaceMembers(wsId);
  const { data: orgMembers = [] } = useFlowTeam();

  const [showInvite, setShowInvite] = useState(false);

  const adminCount = members.filter((m) => m.role === 'admin').length;

  const existingUserIds = useMemo(() => new Set(members.map((m) => String(m.userId))), [members]);

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={t('workspaceManager.pageTitle')} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users size={14} style={{ color: 'var(--flow-text-muted)' }} />
            <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
              {t('workspaceManager.title')}
            </h1>
          </div>
          <button
            type="button"
            onClick={() => setShowInvite(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-[11px] font-semibold"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            <Plus size={11} /> {t('workspaceManager.addMember')}
          </button>
        </div>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-dim)' }}>
          {t('workspaceManager.memberCount', { count: members.length })}
        </p>
      </div>

      <AddWorkspaceMemberDialog
        open={showInvite}
        onOpenChange={setShowInvite}
        workspaceId={wsId}
        orgMembers={orgMembers}
        existingUserIds={existingUserIds}
      />

      <div className="p-6 max-w-2xl space-y-3">
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
          <EmptyState
            icon={Users}
            title={t('workspaceManager.noMembersEmpty')}
            action={{
              label: t('workspaceManager.addMember'),
              onClick: () => setShowInvite(true),
            }}
          />
        ) : (
          <div className="space-y-2 group">
            {members.map((m) => (
              <MemberRow
                key={m.id}
                member={m}
                workspaceId={wsId}
                isLastAdmin={m.role === 'admin' && adminCount <= 1}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default WorkspaceManager;
