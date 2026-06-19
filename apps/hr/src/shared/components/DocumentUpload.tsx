import React, { useCallback, useRef, useState } from 'react';
import { uploadAsset } from '@voxire/api';
import { Upload, FileText } from 'lucide-react';
import { ACCENT } from './ui';

interface Props {
  accept?: string;
  onFile: (file: { name: string; sizeKb: number; type: string; url: string }) => void;
  hint?: string;
}

const DocumentUpload: React.FC<Props> = ({
  accept = '.pdf,.doc,.docx,.png,.jpg',
  onFile,
  hint,
}) => {
  const [dragOver, setDragOver] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const ref = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    async (f: File) => {
      setUploading(true);
      setError(null);
      try {
        const uploaded = await uploadAsset(f);
        onFile({
          name: f.name,
          sizeKb: Math.round((uploaded.size || f.size) / 1024),
          type: uploaded.mime || f.type || 'application/octet-stream',
          url: uploaded.url,
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Upload failed');
      } finally {
        setUploading(false);
        if (ref.current) ref.current.value = '';
      }
    },
    [onFile]
  );

  return (
    <div
      onDragOver={(e) => {
        e.preventDefault();
        setDragOver(true);
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={(e) => {
        e.preventDefault();
        setDragOver(false);
        const f = e.dataTransfer.files?.[0];
        if (f) void handleFile(f);
      }}
      onClick={() => {
        if (!uploading) ref.current?.click();
      }}
      style={{
        padding: 24,
        borderRadius: 12,
        border: `1.5px dashed ${dragOver ? ACCENT : 'var(--hr-border)'}`,
        background: dragOver ? `${ACCENT}11` : 'var(--hr-card)',
        textAlign: 'center',
        cursor: uploading ? 'wait' : 'pointer',
        transition: 'background 0.15s, border-color 0.15s',
      }}
    >
      <div
        style={{
          display: 'inline-flex',
          padding: 12,
          borderRadius: 12,
          background: `${ACCENT}22`,
          color: ACCENT,
          marginBottom: 10,
        }}
      >
        {dragOver || uploading ? <Upload size={20} /> : <FileText size={20} />}
      </div>
      <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
        {uploading ? 'Uploading file...' : 'Drag and drop file here'}
      </div>
      <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 4 }}>
        {hint ?? `or click to browse (${accept})`}
      </div>
      {error && <div style={{ fontSize: 11, color: '#ef4444', marginTop: 8 }}>{error}</div>}
      <input
        ref={ref}
        type="file"
        accept={accept}
        style={{ display: 'none' }}
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f) void handleFile(f);
        }}
      />
    </div>
  );
};

export default DocumentUpload;
