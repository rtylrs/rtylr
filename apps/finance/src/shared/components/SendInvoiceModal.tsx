import React from 'react';
import { Send } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import Modal from './Modal';
import { primaryBtn, ghostBtn } from '../lib/tokens';
import { useSendInvoice } from '../hooks/queries/useInvoices';

interface SendInvoiceModalProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  onSent?: () => void;
}

const SendInvoiceModal: React.FC<SendInvoiceModalProps> = ({
  open,
  onClose,
  invoiceId,
  onSent,
}) => {
  const { t } = useTranslation();
  const { mutate: sendInvoice, isPending: sending } = useSendInvoice();

  const handleSend = (): void => {
    sendInvoice(
      { invoiceId },
      {
        onSuccess: () => {
          onSent?.();
          onClose();
        },
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      blockClose={sending}
      title={t('invoices.sendTitle', { id: invoiceId })}
      width={520}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose} disabled={sending}>
            {t('common.cancel')}
          </button>
          <button type="button" style={primaryBtn} onClick={handleSend} disabled={sending}>
            <Send size={13} />
            {sending ? t('invoices.sending') : t('invoices.sendCta')}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <p style={{ margin: 0, color: 'var(--finance-invoice-muted)' }}>{t('invoices.sendInfo')}</p>
      </div>
    </Modal>
  );
};

export default SendInvoiceModal;
