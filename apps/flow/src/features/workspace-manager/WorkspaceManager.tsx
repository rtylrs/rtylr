import React, { useState, useMemo } from 'react';
import { motion } from 'motion/react';
import { Users, Plus, Trash2, Shield, Eye, Crown, Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useParams } from '@tanstack/react-router';
import {
  useWorkspaceMembers,
  useAddWorkspaceMember,
  useRemoveWorkspaceMember,
  useUpdateWorkspaceMemberRole,
} from './useWorkspaceManager';
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
        onClick={() => {
          if (isLastAdmin) return toast.error(t('toasts.workspaceMember.cannotRemoveLastAdmin'));
          setConfirmRemove(true);
        }}
        className="h-7 w-7 rounded flex items-center justify-center shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
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
  const addMember = useAddWorkspaceMember(wsId);
  const { data: orgMembers = [] } = useFlowTeam();

  const [selectedUserId, setSelectedUserId] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const [showInvite, setShowInvite] = useState(false);
  const [search, setSearch] = useState('');

  const adminCount = members.filter((m) => m.role === 'admin').length;

  const existingUserIds = useMemo(() => new Set(members.map((m) => String(m.userId))), [members]);

  const availableMembers = useMemo(
    () =>
      orgMembers.filter(
        (m) =>
          !existingUserIds.has(String(m.id)) &&
          (search === '' ||
            m.name.toLowerCase().includes(search.toLowerCase()) ||
            m.email.toLowerCase().includes(search.toLowerCase()))
      ),
    [orgMembers, existingUserIds, search]
  );

  const handleAddMember = (): void => {
    if (!selectedUserId) {
      toast.error(t('toasts.workspaceMember.selectMember'));
      return;
    }
    addMember.mutate(
      { workspace_id: Number(wsId), user_id: Number(selectedUserId), role: inviteRole },
      {
        onSuccess: () => {
          toast.success(t('toasts.workspaceMember.added'));
          setSelectedUserId('');
          setSearch('');
          setShowInvite(false);
        },
        onError: () => toast.error(t('toasts.workspaceMember.addFailed')),
      }
    );
  };

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
            onClick={() => setShowInvite((v) => !v)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            <Plus size={11} /> {t('workspaceManager.addMember')}
          </button>
        </div>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-dim)' }}>
          {t('workspaceManager.memberCount', { count: members.length })}
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
              {t('workspaceManager.addMemberHeading')}
            </p>
            <div
              className="flex items-center gap-2 rounded px-3 py-2"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <Search size={12} style={{ color: 'var(--flow-text-dim)' }} />
              <input
                type="text"
                placeholder={t('workspaceManager.searchPlaceholder')}
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setSelectedUserId('');
                }}
                className="flex-1 bg-transparent outline-none text-[12px]"
                style={{ color: 'var(--flow-text)' }}
              />
            </div>
            {search && availableMembers.length > 0 && (
              <div
                className="rounded-lg overflow-hidden divide-y divide-[var(--flow-border)] max-h-48 overflow-y-auto"
                style={{ border: '1px solid var(--flow-border)' }}
              >
                {availableMembers.map((m) => (
                  <button
                    key={m.id}
                    type="button"
                    onClick={() => {
                      setSelectedUserId(String(m.id));
                      setSearch(m.name || m.email);
                    }}
                    className="w-full flex items-center gap-3 px-3 py-2.5 text-left hover:bg-[var(--flow-hover)] transition-colors"
                    style={{
                      background:
                        selectedUserId === String(m.id) ? 'var(--flow-active)' : 'var(--flow-card)',
                    }}
                  >
                    <span
                      className="h-6 w-6 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                      style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
                    >
                      {(m.name || m.email).charAt(0).toUpperCase()}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p
                        className="text-[12px] font-medium truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {m.name}
                      </p>
                      <p className="text-[10px] truncate" style={{ color: 'var(--flow-text-dim)' }}>
                        {m.email}
                      </p>
                    </div>
                  </button>
                ))}
              </div>
            )}
            {search && availableMembers.length === 0 && (
              <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
                {t('workspaceManager.noOrgMembersFound')}
              </p>
            )}
            <div className="flex items-center gap-2">
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member' | 'viewer')}
                className="rounded px-3 py-2 text-[12px] outline-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              >
                {WORKSPACE_ROLES.map((role) => (
                  <option key={role} value={role}>
                    {t(roleLabel(role))}
                  </option>
                ))}
              </select>
              <button
                onClick={handleAddMember}
                disabled={!selectedUserId || addMember.isPending}
                className="flex-1 px-4 py-2 rounded text-[12px] font-semibold disabled:opacity-40"
                style={{ background: 'var(--accent)', color: '#0a0a0a' }}
              >
                {addMember.isPending
                  ? t('workspaceManager.adding')
                  : t('workspaceManager.addToWorkspace')}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowInvite(false);
                  setSearch('');
                  setSelectedUserId('');
                }}
                className="px-3 py-2 rounded text-[12px]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('common.cancel')}
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
            {t('workspaceManager.noMembersEmpty')}
          </div>
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
