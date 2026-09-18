import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, X } from 'lucide-react';
import { Textarea } from './ui';

interface Props {
  onApprove: (note?: string) => void;
  onReject: (note?: string) => void;
  allowNote?: boolean;
  size?: 'sm' | 'md';
}

const ApprovalButtons: React.FC<Props> = ({
  onApprove,
  onReject,
  allowNote = false,
  size = 'sm',
}) => {
  const { t } = useTranslation();
  const [note, setNote] = useState('');
  const [open, setOpen] = useState(false);
  const pad = size === 'md' ? '8px 14px' : '5px 10px';
  const font = size === 'md' ? 12 : 11;

  if (allowNote && open) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <Textarea
          placeholder={t('common.optionalNote', { defaultValue: 'Optional note' })}
          value={note}
          onChange={(e) => setNote(e.target.value)}
          style={{ minHeight: 60 }}
        />
        <div style={{ display: 'flex', gap: 4 }}>
          <button
            type="button"
            onClick={() => {
              onApprove(note);
              setNote('');
              setOpen(false);
            }}
            style={{
              flex: 1,
              padding: pad,
              borderRadius: 8,
              border: '1px solid rgba(16,185,129,0.35)',
              background: 'rgba(16,185,129,0.12)',
              color: '#10b981',
              fontSize: font,
              fontWeight: 600,
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              justifyContent: 'center',
            }}
          >
            <Check size={12} /> {t('common.approve', { defaultValue: 'Approve' })}
          </button>
          <button
            type="button"
            onClick={() => {
              onReject(note);
              setNote('');
              setOpen(false);
            }}
            style={{
              flex: 1,
              padding: pad,
              borderRadius: 8,
              border: '1px solid rgba(239,68,68,0.35)',
              background: 'rgba(239,68,68,0.12)',
              color: '#ef4444',
              fontSize: font,
              fontWeight: 600,
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              justifyContent: 'center',
            }}
          >
            <X size={12} /> {t('common.reject', { defaultValue: 'Reject' })}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', gap: 4 }}>
      <button
        type="button"
        onClick={() => (allowNote ? setOpen(true) : onApprove())}
        style={{
          padding: pad,
          borderRadius: 8,
          border: '1px solid rgba(16,185,129,0.35)',
          background: 'rgba(16,185,129,0.12)',
          color: '#10b981',
          fontSize: font,
          fontWeight: 600,
          cursor: 'pointer',
          display: 'inline-flex',
          alignItems: 'center',
          gap: 4,
        }}
      >
        <Check size={12} /> {t('common.approve', { defaultValue: 'Approve' })}
      </button>
      <button
        type="button"
        onClick={() => (allowNote ? setOpen(true) : onReject())}
        style={{
          padding: pad,
          borderRadius: 8,
          border: '1px solid rgba(239,68,68,0.35)',
          background: 'rgba(239,68,68,0.12)',
          color: '#ef4444',
          fontSize: font,
          fontWeight: 600,
          cursor: 'pointer',
          display: 'inline-flex',
          alignItems: 'center',
          gap: 4,
        }}
      >
        <X size={12} /> {t('common.reject', { defaultValue: 'Reject' })}
      </button>
    </div>
  );
};

export default ApprovalButtons;
