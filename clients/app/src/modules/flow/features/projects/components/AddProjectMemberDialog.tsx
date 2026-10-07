import React, { useEffect, useId, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@voxire/ui';
import { useInviteProjectMember } from '../useProjectMembers';
import { projectRoleLabel } from '../../../shared/lib/roleUtils';
import type { ApiProjectMember } from '../../../shared/services/flowApi';
import type { FlowTeamMember } from '../../../shared/hooks/queries/useFlowTeam';

const PROJECT_MEMBER_ROLES: Array<'admin' | 'member' | 'viewer'> = ['admin', 'member', 'viewer'];

interface AddProjectMemberDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  projectId: string;
  availableMembers: FlowTeamMember[];
  projectMembers: ApiProjectMember[];
}

// Radix Dialog wrapper for inviting a member to a project - matches the modal
// pattern used by every other creation flow in Flow (New project, New task,
// Request time off) instead of the inline expand this replaced. Same
// email + role invite call as before, just in a real modal.
const AddProjectMemberDialog: React.FC<AddProjectMemberDialogProps> = ({
  open,
  onOpenChange,
  projectId,
  availableMembers,
  projectMembers,
}) => {
  const { t } = useTranslation();
  const inviteMember = useInviteProjectMember(projectId);
  const listId = useId();

  const [email, setEmail] = useState('');
  const [role, setRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

  const candidates = useMemo(() => {
    const existingIds = new Set(projectMembers.map((member) => member.userId));
    const existingEmails = new Set(projectMembers.map((member) => member.email.toLowerCase()));

    return availableMembers.filter(
      (member) =>
        member.isActive &&
        member.email &&
        !existingIds.has(member.id) &&
        !existingEmails.has(member.email.toLowerCase())
    );
  }, [availableMembers, projectMembers]);
  const query = email.trim().toLowerCase();
  const matches = useMemo(() => {
    if (!query) return [];
    return candidates
      .filter((member) => `${member.name} ${member.email}`.toLowerCase().includes(query))
      .slice(0, 6);
  }, [candidates, query]);
  const selectedMember = candidates.find((member) => member.email.toLowerCase() === query);

  // Reset on each open so reopening a previously closed dialog doesn't carry
  // over a stale email/role.
  useEffect(() => {
    if (open) {
      setEmail('');
      setRole('member');
      setIsSearchOpen(false);
      setActiveIndex(0);
    }
  }, [open]);

  const handleOpenChange = (nextOpen: boolean): void => {
    if (inviteMember.isPending && !nextOpen) return;
    onOpenChange(nextOpen);
  };

  const handleInvite = (): void => {
    if (inviteMember.isPending) return;
    if (!email.trim()) {
      toast.error(t('flowToasts.projectMember.enterEmail'));
      return;
    }
    if (!selectedMember) {
      toast.error(t('flowToasts.projectMember.selectWorkspaceMember'));
      return;
    }
    inviteMember.mutate(
      { email: selectedMember.email, role },
      {
        onSuccess: () => {
          onOpenChange(false);
        },
      }
    );
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
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
          <div className="relative min-w-0 flex-1">
            <input
              autoFocus
              type="text"
              inputMode="email"
              autoComplete="off"
              role="combobox"
              aria-autocomplete="list"
              aria-expanded={isSearchOpen && matches.length > 0}
              aria-controls={isSearchOpen && matches.length > 0 ? listId : undefined}
              aria-activedescendant={
                isSearchOpen && matches[activeIndex] ? `${listId}-option-${activeIndex}` : undefined
              }
              aria-label={t('flow.projectSettings.emailPlaceholder')}
              placeholder={t('flow.projectSettings.emailPlaceholder')}
              value={email}
              disabled={inviteMember.isPending}
              onFocus={() => setIsSearchOpen(true)}
              onBlur={() => setIsSearchOpen(false)}
              onChange={(e) => {
                setEmail(e.target.value);
                setActiveIndex(0);
                setIsSearchOpen(true);
              }}
              onKeyDown={(e) => {
                if (e.key === 'ArrowDown' && matches.length) {
                  e.preventDefault();
                  setIsSearchOpen(true);
                  setActiveIndex((index) => (index + 1) % matches.length);
                } else if (e.key === 'ArrowUp' && matches.length) {
                  e.preventDefault();
                  setIsSearchOpen(true);
                  setActiveIndex((index) => (index - 1 + matches.length) % matches.length);
                } else if (e.key === 'Escape') {
                  setIsSearchOpen(false);
                } else if (e.key === 'Enter') {
                  e.preventDefault();
                  if (isSearchOpen && matches[activeIndex]) {
                    setEmail(matches[activeIndex].email);
                    setIsSearchOpen(false);
                  } else {
                    handleInvite();
                  }
                }
              }}
              className="w-full rounded-sm px-2 py-1.5 text-flow-title md:text-flow-meta"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
            {isSearchOpen && query && matches.length > 0 && (
              <div
                id={listId}
                role="listbox"
                aria-label={t('flow.projectSettings.memberSuggestionsLabel')}
                className="absolute left-0 right-0 top-full z-50 mt-1 max-h-60 overflow-auto rounded-md p-1 shadow-lg"
                style={{
                  background: 'var(--flow-card)',
                  border: '1px solid var(--flow-border)',
                }}
              >
                {matches.map((member, index) => (
                  <div
                    id={`${listId}-option-${index}`}
                    key={member.id}
                    role="option"
                    aria-selected={index === activeIndex}
                    onMouseDown={(event) => event.preventDefault()}
                    onMouseEnter={() => setActiveIndex(index)}
                    onClick={() => {
                      setEmail(member.email);
                      setIsSearchOpen(false);
                    }}
                    className="cursor-pointer rounded-sm px-2 py-2"
                    style={{
                      background: index === activeIndex ? 'var(--flow-hover)' : 'transparent',
                      color: 'var(--flow-text)',
                    }}
                  >
                    <span className="block truncate text-flow-meta font-medium">{member.name}</span>
                    <span
                      className="block truncate text-flow-meta"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {member.email}
                    </span>
                  </div>
                ))}
              </div>
            )}
            {isSearchOpen && query && matches.length === 0 && (
              <p
                role="status"
                className="absolute left-0 right-0 top-full z-50 mt-1 rounded-md px-3 py-2 text-flow-meta shadow-lg"
                style={{
                  background: 'var(--flow-card)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text-muted)',
                }}
              >
                {t('flow.projectSettings.noMemberMatches')}
              </p>
            )}
          </div>
          <select
            value={role}
            disabled={inviteMember.isPending || !selectedMember}
            onChange={(e) => setRole(e.target.value as 'admin' | 'member' | 'viewer')}
            aria-label={t('flow.projectSettings.roleLabel')}
            className="rounded-sm px-2 py-1.5 text-flow-title md:text-flow-meta"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          >
            {PROJECT_MEMBER_ROLES.map((r) => (
              <option key={r} value={r}>
                {t(projectRoleLabel(r))}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={handleInvite}
            disabled={inviteMember.isPending || !selectedMember}
            className="min-h-[44px] px-3 rounded-sm text-flow-meta font-semibold disabled:opacity-40 md:min-h-0 md:py-1.5"
            style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
          >
            {t('flow.projectSettings.invite')}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AddProjectMemberDialog;
