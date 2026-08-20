import React, { useEffect, useMemo, useState } from 'react';
import { Search, UserSearch } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Dialog, DialogContent, DialogHeader, DialogTitle, EmptyState } from '@voxire/ui';
import { useAddWorkspaceMember } from './useWorkspaceManager';
import type { FlowTeamMember } from '../../shared/hooks/queries/useFlowTeam';
import { roleLabel } from '../../shared/lib/roleUtils';

const WORKSPACE_ROLES: Array<'admin' | 'member' | 'viewer'> = ['admin', 'member', 'viewer'];

interface AddWorkspaceMemberDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  workspaceId: string;
  orgMembers: FlowTeamMember[];
  existingUserIds: Set<string>;
}

// Radix Dialog wrapper for adding a member to a workspace - matches the
// modal pattern used by every other creation flow in Flow (New project,
// New task, Request time off) instead of the inline expand this replaced.
// See docs on ConfirmDialog / RequestLeaveDialog for the same shape.
const AddWorkspaceMemberDialog: React.FC<AddWorkspaceMemberDialogProps> = ({
  open,
  onOpenChange,
  workspaceId,
  orgMembers,
  existingUserIds,
}) => {
  const { t } = useTranslation();
  const addMember = useAddWorkspaceMember(workspaceId);

  const [selectedUserId, setSelectedUserId] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const [search, setSearch] = useState('');

  // Reset on each open so reopening a previously closed dialog doesn't
  // carry over a stale search/selection.
  useEffect(() => {
    if (open) {
      setSelectedUserId('');
      setInviteRole('member');
      setSearch('');
    }
  }, [open]);

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
      { workspace_id: Number(workspaceId), user_id: Number(selectedUserId), role: inviteRole },
      {
        onSuccess: () => {
          toast.success(t('toasts.workspaceMember.added'));
          onOpenChange(false);
        },
        onError: () => toast.error(t('toasts.workspaceMember.addFailed')),
      }
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-md"
        style={{ background: 'var(--flow-card)', color: 'var(--flow-text)' }}
      >
        <DialogHeader>
          <DialogTitle style={{ color: 'var(--flow-text)' }}>
            {t('workspaceManager.addMemberHeading')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-3">
          <div
            className="flex items-center gap-2 rounded px-3 py-2"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <Search size={12} style={{ color: 'var(--flow-text-dim)' }} />
            <input
              autoFocus
              type="text"
              aria-label={t('workspaceManager.searchPlaceholder')}
              placeholder={t('workspaceManager.searchPlaceholder')}
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setSelectedUserId('');
              }}
              className="flex-1 bg-transparent outline-none text-[16px] md:text-[12px]"
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
            <EmptyState
              icon={UserSearch}
              title={t('workspaceManager.noOrgMembersFound')}
              action={{
                label: t('workspaceManager.clearSearchCta'),
                onClick: () => setSearch(''),
              }}
              size="compact"
            />
          )}

          <div className="flex items-center gap-2">
            <select
              value={inviteRole}
              onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member' | 'viewer')}
              aria-label={t('workspaceManager.roleLabel')}
              className="min-h-[44px] rounded px-3 text-[16px] outline-none md:min-h-0 md:py-2 md:text-[12px]"
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
              type="button"
              onClick={handleAddMember}
              disabled={!selectedUserId || addMember.isPending}
              className="flex-1 min-h-[44px] px-4 rounded text-[13px] font-semibold disabled:opacity-40 md:min-h-0 md:py-2"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              {addMember.isPending
                ? t('workspaceManager.adding')
                : t('workspaceManager.addToWorkspace')}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AddWorkspaceMemberDialog;
