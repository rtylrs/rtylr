import React from 'react';
import Modal from './Modal';
import { ghostBtn, primaryBtn } from '../lib/tokens';
import { AlertTriangle } from 'lucide-react';

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmLabel?: string;
  destructive?: boolean;
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel = 'Confirm',
  destructive,
}) => (
  <Modal
    open={open}
    onClose={onClose}
    title={title}
    width={400}
    footer={
      <>
        <button style={ghostBtn} onClick={onClose}>
          Cancel
        </button>
        <button
          style={{ ...primaryBtn, background: destructive ? '#ef4444' : primaryBtn.background }}
          onClick={() => {
            onConfirm();
            onClose();
          }}
        >
          {confirmLabel}
        </button>
      </>
    }
  >
    <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: 36,
          height: 36,
          borderRadius: 10,
          background: destructive ? 'rgba(239,68,68,0.12)' : 'rgba(245,158,11,0.12)',
          color: destructive ? '#ef4444' : '#f59e0b',
          flexShrink: 0,
        }}
      >
        <AlertTriangle size={18} />
      </div>
      <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
        {message}
      </p>
    </div>
  </Modal>
);

export default ConfirmDialog;
