import React from 'react';
import { Send } from 'lucide-react';
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
      title={`Send ${invoiceId}`}
      width={520}
      footer={
        <>
          <button style={ghostBtn} onClick={onClose} disabled={sending}>
            Cancel
          </button>
          <button type="button" style={primaryBtn} onClick={handleSend} disabled={sending}>
            <Send size={13} />
            {sending ? 'Sending...' : 'Send Invoice'}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <p style={{ margin: 0, color: 'var(--finance-invoice-muted)' }}>
          This invoice will be sent to the recipient configured for it.
        </p>
      </div>
    </Modal>
  );
};

export default SendInvoiceModal;
