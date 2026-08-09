import React from 'react';
import { AlertTriangle, X } from 'lucide-react';

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
  if (!open) return null;
  return (
    <div
      role="dialog"
      aria-modal="true"
      data-testid="confirm-modal"
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'rgba(0,0,0,0.6)' }}
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
              background: danger ? 'rgba(248,113,113,0.1)' : 'rgba(245,158,11,0.1)',
            }}
          >
            <AlertTriangle size={16} style={{ color: danger ? '#f87171' : '#f59e0b' }} />
          </div>
          <div className="flex-1">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                {title}
              </h3>
              <button onClick={onCancel}>
                <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
              </button>
            </div>
            {description && (
              <p className="text-xs mt-1" style={{ color: 'var(--hub-text-muted)' }}>
                {description}
              </p>
            )}
          </div>
        </div>
        <div className="flex gap-2 justify-end mt-4">
          <button
            onClick={onCancel}
            className="px-3 py-1.5 rounded-lg text-xs font-medium"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            {cancelLabel}
          </button>
          <button
            data-testid="confirm-modal-confirm"
            onClick={onConfirm}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{
              background: danger ? '#f87171' : '#f59e0b',
              color: '#000',
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
