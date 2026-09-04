import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@voxire/ui';
import { useInviteProjectMember } from '../useProjectMembers';
import { roleLabel } from '../../../shared/lib/roleUtils';

const PROJECT_MEMBER_ROLES: Array<'admin' | 'member' | 'viewer'> = ['admin', 'member', 'viewer'];

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
          <DialogTitle style={{ color: 'var(--flow-text)' }}>
            {t('flow.projectSettings.addMemberHeading')}
          </DialogTitle>
          <DialogDescription style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.projectSettings.addMemberDescription')}
          </DialogDescription>
        </DialogHeader>

        <div className="flex gap-2">
          <input
            autoFocus
            type="email"
            aria-label={t('flow.projectSettings.emailPlaceholder')}
            placeholder={t('flow.projectSettings.emailPlaceholder')}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleInvite();
              }
            }}
            className="flex-1 rounded-sm px-2 py-1.5 text-title md:text-meta outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          />
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as 'admin' | 'member' | 'viewer')}
            aria-label={t('flow.projectSettings.roleLabel')}
            className="rounded-sm px-2 py-1.5 text-title md:text-meta outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          >
            {PROJECT_MEMBER_ROLES.map((r) => (
              <option key={r} value={r}>
                {t(roleLabel(r))}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={handleInvite}
            disabled={inviteMember.isPending}
            className="min-h-[44px] px-3 rounded-sm text-meta font-semibold disabled:opacity-40 md:min-h-0 md:py-1.5"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            {t('flow.projectSettings.invite')}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AddProjectMemberDialog;
