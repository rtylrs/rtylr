import React, { useState } from 'react';
import { Send } from 'lucide-react';
import { toast } from 'sonner';
import Modal from './Modal';
import { inputStyle, primaryBtn, ghostBtn, labelStyle } from '../lib/tokens';
import { useSendInvoice } from '../hooks/queries/useInvoices';

interface SendInvoiceModalProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  defaultTo?: string;
  defaultSubject?: string;
  defaultBody?: string;
  onSent?: () => void;
}

const SendInvoiceModal: React.FC<SendInvoiceModalProps> = ({
  open,
  onClose,
  invoiceId,
  defaultTo = '',
  defaultSubject,
  defaultBody,
  onSent,
}) => {
  const [to, setTo] = useState(defaultTo);
  const [cc, setCc] = useState('');
  const [subject, setSubject] = useState(defaultSubject ?? `Invoice ${invoiceId} from rtylr`);
  const [body, setBody] = useState(
    defaultBody ??
      `Hi,\n\nPlease find invoice ${invoiceId} attached. Payment is due upon receipt.\n\nThank you,\nrtylr`
  );

  const { mutate: sendInvoice, isPending: sending } = useSendInvoice();

  const handleSend = (): void => {
    if (!to.trim()) {
      toast.error('Recipient required');
      return;
    }
    const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRe.test(to.trim())) {
      toast.error('Please enter a valid email address');
      return;
    }
    if (cc.trim() && !emailRe.test(cc.trim())) {
      toast.error('CC must be a valid email address');
      return;
    }
    sendInvoice(
      { invoiceId, to, cc: cc || undefined, subject, body },
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
          <button style={primaryBtn} onClick={handleSend} disabled={sending}>
            <Send size={13} />
            {sending ? 'Sending...' : 'Send Invoice'}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>To</label>
          <input
            value={to}
            onChange={(e) => setTo(e.target.value)}
            placeholder="client@example.com"
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>CC</label>
          <input
            value={cc}
            onChange={(e) => setCc(e.target.value)}
            placeholder="optional"
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>Subject</label>
          <input value={subject} onChange={(e) => setSubject(e.target.value)} style={inputStyle} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>Message</label>
          <textarea
            rows={6}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
          />
        </div>
      </div>
    </Modal>
  );
};

export default SendInvoiceModal;
