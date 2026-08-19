import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@voxire/ui';
import { useInviteProjectMember } from '../useProjectMembers';

interface AddProjectMemberDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  projectId: string;
}

// Radix Dialog wrapper for inviting a member to a project - matches the modal
// pattern used by every other creation flow in Flow (New project, New task,
// Request time off) instead of the inline expand this replaced. Same
// email + role invite call as before, just in a real modal.
const AddProjectMemberDialog: React.FC<AddProjectMemberDialogProps> = ({
  open,
  onOpenChange,
  projectId,
}) => {
  const { t } = useTranslation();
  const inviteMember = useInviteProjectMember(projectId);

  const [email, setEmail] = useState('');
  const [role, setRole] = useState<'admin' | 'member' | 'viewer'>('member');

  // Reset on each open so reopening a previously closed dialog doesn't carry
  // over a stale email/role.
  useEffect(() => {
    if (open) {
      setEmail('');
      setRole('member');
    }
  }, [open]);

  const handleInvite = (): void => {
    if (!email.trim()) {
      toast.error(t('toasts.projectMember.enterEmail'));
      return;
    }
    inviteMember.mutate(
      { email: email.trim(), role },
      {
        onSuccess: () => {
          onOpenChange(false);
        },
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
          <DialogTitle style={{ color: 'var(--flow-text)' }}>Add member</DialogTitle>
        </DialogHeader>

        <div className="flex gap-2">
          <input
            autoFocus
            type="email"
            placeholder={t('flow.projectSettings.emailPlaceholder')}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleInvite();
              }
            }}
            className="flex-1 rounded px-2 py-1.5 text-[16px] md:text-[11px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
          />
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as 'admin' | 'member' | 'viewer')}
            aria-label={t('flow.projectSettings.roleLabel')}
            className="rounded px-2 py-1.5 text-[16px] md:text-[11px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
          >
            <option value="admin">Admin</option>
            <option value="member">Member</option>
            <option value="viewer">Viewer</option>
          </select>
          <button
            onClick={handleInvite}
            disabled={inviteMember.isPending}
            className="px-3 py-1.5 rounded text-[11px] font-semibold disabled:opacity-40"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            Invite
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AddProjectMemberDialog;
