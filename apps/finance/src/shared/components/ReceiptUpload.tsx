import React, { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage, uploadAsset } from '@voxire/api';
import { cdnImageUrl } from '@voxire/config';
import { Upload, FileImage, X } from 'lucide-react';

interface ReceiptUploadProps {
  value?: string | null;
  onChange: (dataUrl: string | null) => void;
}

const ReceiptUpload: React.FC<ReceiptUploadProps> = ({ value, onChange }) => {
  const { t } = useTranslation();
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleFile = async (file: File | undefined | null): Promise<void> => {
    if (!file) return;
    setUploading(true);
    setError(null);
    try {
      const uploaded = await uploadAsset(file);
      onChange(uploaded.url);
    } catch (err) {
      setError(getOneKitErrorMessage(err, 'Upload failed'));
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  return (
    <div
      onDragOver={(e) => {
        e.preventDefault();
        setDragging(true);
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={(e) => {
        e.preventDefault();
        setDragging(false);
        void handleFile(e.dataTransfer.files?.[0]);
      }}
      onClick={() => {
        if (!uploading) fileRef.current?.click();
      }}
      style={{
        border: `2px dashed ${dragging ? 'var(--finance-accent)' : 'var(--hub-border)'}`,
        borderRadius: 12,
        padding: 24,
        textAlign: 'center',
        cursor: uploading ? 'wait' : 'pointer',
        background: dragging ? 'var(--finance-accent-subtle)' : 'var(--hub-hover)',
        transition: 'all 0.15s',
        position: 'relative',
        minHeight: 160,
      }}
    >
      <input
        ref={fileRef}
        type="file"
        accept="image/*"
        style={{ display: 'none' }}
        onChange={(e) => {
          void handleFile(e.target.files?.[0]);
        }}
      />
      {uploading ? (
        <>
          <Upload size={28} style={{ color: 'var(--finance-accent)', marginBottom: 10 }} />
          <div style={{ fontSize: 13, color: 'var(--hub-text)', fontWeight: 600 }}>
            {t('expenses.uploadingReceipt')}
          </div>
        </>
      ) : value ? (
        <div style={{ position: 'relative' }}>
          <img
            src={cdnImageUrl(value)}
            alt={t('common.receipt')}
            style={{ maxHeight: 180, maxWidth: '100%', borderRadius: 8 }}
          />
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onChange(null);
            }}
            style={{
              position: 'absolute',
              top: -8,
              insetInlineEnd: -8,
              background: 'var(--finance-danger)',
              color: 'var(--finance-white)',
              border: 'none',
              borderRadius: '50%',
              width: 22,
              height: 22,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <X size={12} />
          </button>
        </div>
      ) : (
        <>
          <Upload size={28} style={{ color: 'var(--hub-text-dim)', marginBottom: 10 }} />
          <div style={{ fontSize: 13, color: 'var(--hub-text)', fontWeight: 600 }}>
            {t('expenses.dropReceipt')}
          </div>
          <div style={{ fontSize: 11, color: 'var(--hub-text-dim)', marginTop: 6 }}>
            {t('expenses.receiptFormats')}
          </div>
        </>
      )}
      {value && (
        <div
          style={{
            marginTop: 10,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 6,
            color: 'var(--hub-text-dim)',
            fontSize: 11,
          }}
        >
          <FileImage size={12} /> Click to replace
        </div>
      )}
      {error && (
        <div style={{ marginTop: 10, color: 'var(--finance-danger)', fontSize: 11 }}>{error}</div>
      )}
    </div>
  );
};

export default ReceiptUpload;
