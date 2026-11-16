import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@voxire/ui';

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
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle className="text-flow-section font-semibold">
            {t('flow.status.newStatusTitle')}
          </DialogTitle>
          <DialogDescription className="text-flow-meta">
            {t('flow.status.newStatusHint')}
          </DialogDescription>
        </DialogHeader>
        <form
          className="mt-2"
          onSubmit={(e) => {
            e.preventDefault();
            submit();
          }}
        >
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
          <div className="mt-4 flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="rounded-lg px-3 py-1.5 text-flow-meta font-medium"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.status.cancel')}
            </button>
            <button
              type="submit"
              disabled={!name.trim() || pending}
              className="rounded-lg px-3 py-1.5 text-flow-meta font-semibold disabled:opacity-40"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              {t('flow.status.addStatus')}
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default NewStatusDialog;
