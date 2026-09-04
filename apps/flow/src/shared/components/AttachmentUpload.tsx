import React, { useRef, useState } from 'react';
import { uploadAsset } from '@voxire/api';
import { Upload, FileText, Image as ImageIcon, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Attachment } from '../stores/flowStore';

interface AttachmentUploadProps {
  attachments: Attachment[];
  onAdd?: (name: string, size: number, mime: string, url: string) => void;
  onRemove?: (id: string) => void;
  disabled?: boolean;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export const AttachmentUpload: React.FC<AttachmentUploadProps> = ({
  attachments,
  onAdd,
  onRemove,
  disabled,
}) => {
  const { t } = useTranslation();
  const ref = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setError(null);
    try {
      const uploaded = await uploadAsset(file);
      onAdd?.(file.name, uploaded.size || file.size, uploaded.mime || file.type, uploaded.url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  return (
    <div>
      <div className="space-y-1.5 mb-2">
        {attachments.map((a) => {
          const Icon = a.mime.startsWith('image') ? ImageIcon : FileText;
          return (
            <div
              key={a.id}
              className="flex items-center gap-2 px-2.5 py-2 rounded-lg group"
              style={{
                background: 'var(--flow-hover)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <Icon size={14} style={{ color: 'var(--flow-text-muted)' }} />
              <div className="flex-1 min-w-0">
                <p className="text-meta font-medium truncate" style={{ color: 'var(--flow-text)' }}>
                  {a.name}
                </p>
                <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {formatSize(a.size)} · {a.uploadedBy}
                </p>
              </div>
              {onRemove && !disabled && (
                <button
                  type="button"
                  onClick={() => onRemove(a.id)}
                  aria-label={t('flow.taskDetail.removeAttachmentLabel')}
                  className="h-5 w-5 rounded-sm opacity-0 group-hover:opacity-100 flex items-center justify-center hover:bg-[var(--flow-active)]"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  <X size={12} />
                </button>
              )}
            </div>
          );
        })}
      </div>
      <button
        type="button"
        onClick={() => {
          if (!uploading && !disabled) ref.current?.click();
        }}
        disabled={uploading || disabled}
        className="flex items-center gap-2 px-3 py-2 rounded-lg text-meta w-full transition-colors disabled:opacity-50"
        style={{
          background: 'var(--flow-input-bg)',
          border: '1px dashed var(--flow-border-control)',
          color: 'var(--flow-text-muted)',
          cursor: uploading ? 'wait' : disabled ? 'not-allowed' : 'pointer',
        }}
      >
        <Upload size={12} />
        {uploading ? 'Uploading...' : 'Add attachment'}
      </button>
      {error && (
        <p className="mt-1 text-meta" style={{ color: 'var(--flow-danger-fg)' }}>
          {error}
        </p>
      )}
      <input
        ref={ref}
        type="file"
        className="hidden"
        onChange={(event) => {
          void handleFile(event);
        }}
      />
    </div>
  );
};

export default AttachmentUpload;
