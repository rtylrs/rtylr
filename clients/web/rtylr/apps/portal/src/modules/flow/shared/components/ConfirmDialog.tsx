import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@voxire/ui';

// Drop-in replacement for window.confirm() inside Flow. Two flavors:
//   - basic: just shows the prompt, primary button enabled immediately
//   - require-typed: user must retype `requireText` to enable the
//     destructive button (use for project / workspace deletion).
//
// Stays controlled so callers don't need to thread their own open
// state - they just call useConfirm() and `await confirm(...)`.

export interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: React.ReactNode;
  /**
   * When set, the user has to retype this exact string (whitespace-trimmed)
   * to enable the confirm button. Used for higher-friction destructive
   * actions (delete project, delete workspace).
   */
  requireText?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  pending?: boolean;
  onConfirm: () => void;
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  onOpenChange,
  title,
  description,
  requireText,
  confirmLabel,
  cancelLabel,
  pending,
  onConfirm,
}) => {
  const { t } = useTranslation();
  const typedInputId = useId();
  const [typed, setTyped] = useState('');

  // Reset typed-confirmation on each open so reopening a previously
  // closed dialog doesn't carry over stale input.
  useEffect(() => {
    if (open) setTyped('');
  }, [open]);

  const typedOk = !requireText || typed.trim() === requireText;
  const disabled = !typedOk || !!pending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="text-flow-section font-semibold flex items-center gap-2">
            <AlertTriangle size={14} style={{ color: 'var(--flow-danger-fg)' }} />
            {title}
          </DialogTitle>
          <DialogDescription className="text-flow-meta">{description}</DialogDescription>
        </DialogHeader>

        {requireText ? (
          <div className="mt-2">
            {/*
              A visible label, doing two jobs at once: it gives the input an
              accessible name, which it had none of, and it states the
              requirement, which the dialog never did. Before this a
              screen-reader user met an unnamed text field next to a disabled
              destructive button with no stated reason for either, and a sighted
              user met a button that looked broken; the only hint was the
              input's own placeholder.
            */}
            <label
              htmlFor={typedInputId}
              className="block text-flow-meta mb-1"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.confirm.typeToConfirm', { word: requireText })}
            </label>
            <input
              id={typedInputId}
              autoFocus
              type="text"
              value={typed}
              onChange={(e) => setTyped(e.target.value)}
              placeholder={requireText}
              className="w-full rounded-lg px-3 py-2 text-flow-body"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
          </div>
        ) : null}

        <div className="flex justify-end gap-2 mt-4">
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="px-3 py-1.5 rounded-lg text-flow-meta font-medium"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {cancelLabel ?? t('flow.confirm.cancel')}
          </button>
          <button
            type="button"
            disabled={disabled}
            onClick={onConfirm}
            className="px-3 py-1.5 rounded-lg text-flow-meta font-semibold disabled:opacity-40"
            style={{ background: 'var(--flow-danger-fill)', color: 'var(--flow-danger-on-fill)' }}
          >
            {pending ? t('flow.confirm.working') : (confirmLabel ?? t('flow.confirm.delete'))}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default ConfirmDialog;
