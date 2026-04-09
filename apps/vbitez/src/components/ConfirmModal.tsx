import React from 'react';

interface ConfirmModalProps {
  open: boolean;
  title: string;
  message?: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  children?: React.ReactNode;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  open,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  danger,
  onConfirm,
  onCancel,
  children,
}) => {
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      role="dialog"
      aria-modal="true"
    >
      <div className="w-full max-w-md rounded-2xl bg-pos-card border border-pos-border p-5 space-y-4">
        <div>
          <h2 className="text-lg font-bold text-pos-text">{title}</h2>
          {message && <p className="text-sm text-pos-muted mt-1">{message}</p>}
        </div>
        {children}
        <div className="grid grid-cols-2 gap-2">
          <button
            className="h-12 rounded-lg bg-white/10 text-pos-text font-semibold hover:bg-white/20 transition"
            onClick={onCancel}
          >
            {cancelText}
          </button>
          <button
            className={`h-12 rounded-lg font-semibold transition ${
              danger
                ? 'bg-destructive text-white hover:opacity-90'
                : 'bg-pos-text text-pos-bg hover:opacity-90'
            }`}
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
