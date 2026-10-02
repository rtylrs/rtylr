import React, { useId } from 'react';
import { AlertTriangle, X } from 'lucide-react';
import { useDialogFocus } from '../hooks/useDialogFocus';

interface ConfirmModalProps {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  open,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  danger,
  onConfirm,
  onCancel,
}) => {
  const titleId = useId();
  const descriptionId = useId();
  const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose: onCancel });

  if (!open) return null;
  return (
    <div
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      data-testid="confirm-modal"
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'var(--hub-overlay-heavy)' }}
      tabIndex={-1}
      onClick={onCancel}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-sm rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-start gap-3">
          <div
            className="flex items-center justify-center h-10 w-10 rounded-full shrink-0"
            style={{
              background: danger
                ? 'color-mix(in srgb, var(--erp-danger) 10%, transparent)'
                : 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
            }}
          >
            <AlertTriangle
              size={16}
              style={{ color: danger ? 'var(--erp-danger)' : 'var(--erp-accent)' }}
            />
          </div>
          <div className="flex-1">
            <div className="flex items-center justify-between">
              <h3
                id={titleId}
                className="text-sm font-semibold"
                style={{ color: 'var(--hub-text)' }}
              >
                {title}
              </h3>
              <button type="button" onClick={onCancel} aria-label="Close dialog">
                <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
              </button>
            </div>
            {description && (
              <p
                id={descriptionId}
                className="text-xs mt-1"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                {description}
              </p>
            )}
          </div>
        </div>
        <div className="flex gap-2 justify-end mt-4">
          <button
            type="button"
            onClick={onCancel}
            className="px-3 py-1.5 rounded-lg text-xs font-medium"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            data-testid="confirm-modal-confirm"
            onClick={onConfirm}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{
              background: danger ? 'var(--erp-danger)' : 'var(--erp-accent)',
              color: 'var(--dash-text)',
            }}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
