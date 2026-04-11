import React, { useRef } from 'react';
import { Paperclip, Upload, FileText, Image as ImageIcon, X } from 'lucide-react';
import { Attachment } from '../stores/flowStore';

interface AttachmentUploadProps {
  attachments: Attachment[];
  onAdd?: (name: string, size: number, mime: string) => void;
  onRemove?: (id: string) => void;
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
}) => {
  const ref = useRef<HTMLInputElement>(null);

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    onAdd?.(file.name, file.size, file.type);
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
                <p
                  className="text-[12px] font-medium truncate"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {a.name}
                </p>
                <p className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                  {formatSize(a.size)} · {a.uploadedBy}
                </p>
              </div>
              {onRemove && (
                <button
                  onClick={() => onRemove(a.id)}
                  className="h-5 w-5 rounded opacity-0 group-hover:opacity-100 flex items-center justify-center hover:bg-[var(--flow-active)]"
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
        onClick={() => ref.current?.click()}
        className="flex items-center gap-2 px-3 py-2 rounded-lg text-[12px] w-full transition-colors"
        style={{
          background: 'var(--flow-input-bg)',
          border: '1px dashed var(--flow-border)',
          color: 'var(--flow-text-muted)',
        }}
      >
        <Upload size={12} />
        Add attachment
      </button>
      <input ref={ref} type="file" className="hidden" onChange={handleFile} />
    </div>
  );
};

export default AttachmentUpload;
