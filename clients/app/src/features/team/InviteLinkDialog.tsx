import React from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Copy } from 'lucide-react';
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from '../../shared/components/ui';

interface InviteLinkDialogProps {
  /** The activation link the server returned, or null when the dialog is closed. */
  link: string | null;
  email: string;
  /** Whether an invitation email is on its way. False means the link is the only way in. */
  emailQueued: boolean;
  onClose: () => void;
}

/**
 * Shows the activation link for a member who was just invited, so the inviter
 * can pass it on when the email does not arrive. The link is a one-time
 * credential: it is held only in component state and never stored.
 */
export function InviteLinkDialog({
  link,
  email,
  emailQueued,
  onClose,
}: InviteLinkDialogProps): React.JSX.Element {
  const { t } = useTranslation();

  const copy = async (): Promise<void> => {
    if (!link) return;
    try {
      await navigator.clipboard.writeText(link);
      toast.success(t('team.inviteUser.inviteLinkCopied'));
    } catch {
      toast.error(t('team.inviteUser.inviteLinkCopyFailed'));
    }
  };

  return (
    <Dialog open={link !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('team.inviteUser.inviteLinkTitle')}</DialogTitle>
          <DialogDescription>
            {t(
              emailQueued
                ? 'team.inviteUser.inviteLinkDescription'
                : 'team.inviteUser.inviteLinkEmailNotSent',
              { email }
            )}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-2 px-6">
          <Label htmlFor="invite-link">{t('team.inviteUser.inviteLinkLabel')}</Label>
          <div className="flex gap-2">
            <Input
              id="invite-link"
              readOnly
              dir="ltr"
              value={link ?? ''}
              onFocus={(event) => event.currentTarget.select()}
            />
            <Button type="button" variant="outline" onClick={() => void copy()}>
              <Copy size={14} aria-hidden="true" />
              {t('team.inviteUser.copyInviteLink')}
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" onClick={onClose}>
            {t('team.inviteUser.inviteLinkDone')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
