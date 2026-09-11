import React from 'react';
import { useTranslation } from 'react-i18next';
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
}) => {
  const { t } = useTranslation();
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      width={400}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose}>
            {t('common.cancel')}
          </button>
          <button
            type="button"
            style={{
              ...primaryBtn,
              background: destructive ? 'var(--finance-danger)' : primaryBtn.background,
            }}
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
            background: destructive ? 'var(--finance-danger-soft)' : 'var(--finance-warning-soft)',
            color: destructive ? 'var(--finance-danger)' : 'var(--finance-warning)',
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
};

export default ConfirmDialog;
