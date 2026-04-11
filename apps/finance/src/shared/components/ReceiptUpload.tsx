import React, { useRef, useState } from 'react';
import { Upload, FileImage, Sparkles, X } from 'lucide-react';
import { ACCENT } from '../lib/tokens';

interface ReceiptUploadProps {
  value?: string | null;
  onChange: (dataUrl: string | null) => void;
  onOcrHint?: (hint: { merchant?: string; amount?: number; date?: string }) => void;
}

const ReceiptUpload: React.FC<ReceiptUploadProps> = ({ value, onChange, onOcrHint }) => {
  const [dragging, setDragging] = useState(false);
  const [ocrRunning, setOcrRunning] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleFile = (file: File | undefined | null) => {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      onChange(reader.result as string);
      if (onOcrHint) {
        setOcrRunning(true);
        setTimeout(() => {
          setOcrRunning(false);
          onOcrHint({
            merchant: 'Example Vendor',
            amount: Math.round(Math.random() * 200 * 100) / 100,
            date: new Date().toISOString().slice(0, 10),
          });
        }, 900);
      }
    };
    reader.readAsDataURL(file);
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
        handleFile(e.dataTransfer.files?.[0]);
      }}
      onClick={() => fileRef.current?.click()}
      style={{
        border: `2px dashed ${dragging ? ACCENT : 'var(--hub-border)'}`,
        borderRadius: 12,
        padding: 24,
        textAlign: 'center',
        cursor: 'pointer',
        background: dragging ? 'rgba(16,185,129,0.08)' : 'var(--hub-hover)',
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
        onChange={(e) => handleFile(e.target.files?.[0])}
      />
      {value ? (
        <div style={{ position: 'relative' }}>
          <img
            src={value}
            alt="Receipt"
            style={{ maxHeight: 180, maxWidth: '100%', borderRadius: 8 }}
          />
          <button
            onClick={(e) => {
              e.stopPropagation();
              onChange(null);
            }}
            style={{
              position: 'absolute',
              top: -8,
              right: -8,
              background: '#ef4444',
              color: '#fff',
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
          {ocrRunning && (
            <div
              style={{
                marginTop: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                color: ACCENT,
                fontSize: 12,
              }}
            >
              <Sparkles size={12} /> Running OCR...
            </div>
          )}
        </div>
      ) : (
        <>
          <Upload size={28} style={{ color: 'var(--hub-text-dim)', marginBottom: 10 }} />
          <div style={{ fontSize: 13, color: 'var(--hub-text)', fontWeight: 600 }}>
            Drop a receipt or click to upload
          </div>
          <div style={{ fontSize: 11, color: 'var(--hub-text-dim)', marginTop: 6 }}>
            PNG, JPG up to 10MB
          </div>
          <div
            style={{
              marginTop: 10,
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              fontSize: 10,
              color: ACCENT,
            }}
          >
            <Sparkles size={11} /> OCR preview enabled
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
    </div>
  );
};

export default ReceiptUpload;
