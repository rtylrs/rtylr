import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../../shared/components/ui/dialog';
import { Button } from '../../../../../shared/components/ui/button';

interface NewStatusDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called with the trimmed name. The caller closes the dialog once it is created. */
  onCreate: (name: string) => void;
  pending?: boolean;
}

/** Asks for a new status's name; the bucket and place are decided by the caller. */
const NewStatusDialog = ({
  open,
  onOpenChange,
  onCreate,
  pending,
}: NewStatusDialogProps): React.JSX.Element => {
  const { t } = useTranslation();
  const inputId = useId();
  const [name, setName] = useState('');

  useEffect(() => {
    if (open) setName('');
  }, [open]);

  const submit = (): void => {
    const trimmed = name.trim();
    if (trimmed && !pending) onCreate(trimmed);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="portal-flow-dialog max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('flow.status.newStatusTitle')}</DialogTitle>
          <DialogDescription>{t('flow.status.newStatusHint')}</DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            submit();
          }}
        >
          <div className="px-6 pb-6">
            <label
              htmlFor={inputId}
              className="mb-1 block text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.status.statusName')}
            </label>
            <input
              id={inputId}
              autoFocus
              value={name}
              maxLength={100}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('flow.status.newStatusPlaceholder')}
              className="w-full rounded-lg px-3 py-2 text-flow-body"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('flow.status.cancel')}
            </Button>
            <Button type="submit" disabled={!name.trim() || pending}>
              {t('flow.status.addStatus')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default NewStatusDialog;
