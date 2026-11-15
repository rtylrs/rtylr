import React, { useRef, useState } from 'react';
import { uploadAsset } from '@voxire/api';
import { getCdnBaseUrl } from '@voxire/config';
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@voxire/ui';
import { ExternalLink, Upload, FileText, Image as ImageIcon, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Attachment, useFlowStore } from '../stores/flowStore';
import { useResolveAssetUrl, useSignedAssetUrl } from '../hooks/queries/useSignedAssetUrl';

interface AttachmentUploadProps {
  attachments: Attachment[];
  onAdd?: (name: string, size: number, mime: string, url: string) => void;
  onRemove?: (id: string) => void;
  /**
   * Whether the viewer may remove this attachment. The server lets the uploader
   * remove their own and project admins remove any, so the button is offered
   * only where the delete will succeed. Defaults to every attachment.
   */
  canRemove?: (attachment: Attachment) => boolean;
  disabled?: boolean;
  /** Names by user id, for people the business team list does not cover. */
  peopleNames?: Record<string, string>;
}

/**
 * The attachment's address when it is safe to load or link: http(s) only. The
 * URL is stored as given by whoever attached the file, so anything else (a
 * `javascript:` link, say) is not rendered.
 */
export function safeAttachmentUrl(url: string | undefined): string | null {
  if (!url) return null;
  try {
    // A relative address is a path on the CDN, which is where uploads live.
    const parsed = new URL(url, `${getCdnBaseUrl()}/`);
    return parsed.protocol === 'https:' || parsed.protocol === 'http:' ? parsed.href : null;
  } catch {
    return null;
  }
}

// A 4:3 frame holds both wide desktop and tall phone screenshots. The image is
// fitted inside it, not cropped, so the thumbnail shows the whole picture.
const THUMB_FRAME =
  'h-12 w-16 shrink-0 rounded-md overflow-hidden flex items-center justify-center';

const AttachmentThumb: React.FC<{ url: string; name: string }> = ({ url, name }) => {
  const signed = useSignedAssetUrl(url);
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const failed = signed.url !== null && failedSrc === signed.url;
  return (
    <span
      className={THUMB_FRAME}
      style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
    >
      {signed.url && !failed ? (
        <img
          src={signed.url}
          alt={name}
          loading="lazy"
          onError={() => setFailedSrc(signed.url)}
          className="h-full w-full object-contain"
        />
      ) : signed.loading ? null : (
        <ImageIcon size={16} style={{ color: 'var(--flow-text-muted)' }} />
      )}
    </span>
  );
};

const PreviewBody: React.FC<{ url: string; name: string }> = ({ url, name }) => {
  const { t } = useTranslation();
  const signed = useSignedAssetUrl(url);
  if (!signed.url) {
    return <div className="h-[40vh] rounded-md" style={{ background: 'var(--flow-hover)' }} />;
  }
  return (
    <>
      <img
        src={signed.url}
        alt={name}
        className="max-h-[75vh] w-full rounded-md object-contain"
        style={{ background: 'var(--flow-hover)' }}
      />
      <a
        href={signed.url}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1.5 text-flow-meta font-medium hover:underline"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <ExternalLink size={12} />
        {t('flow.taskDetail.openOriginal')}
      </a>
    </>
  );
};

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export const AttachmentUpload: React.FC<AttachmentUploadProps> = ({
  attachments,
  onAdd,
  onRemove,
  canRemove,
  disabled,
  peopleNames,
}) => {
  const { t } = useTranslation();
  const members = useFlowStore((state) => state.members);
  const resolveUrl = useResolveAssetUrl();
  const [preview, setPreview] = useState<{ name: string; url: string } | null>(null);
  const uploaderName = (id: string): string =>
    members.find((member) => String(member.id) === String(id))?.name ??
    peopleNames?.[String(id)] ??
    id;
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
          const isImage = a.mime.startsWith('image');
          const Icon = isImage ? ImageIcon : FileText;
          const href = safeAttachmentUrl(a.url);
          const open = () => {
            if (!href) return;
            if (isImage) {
              setPreview({ name: a.name, url: href });
              return;
            }
            // Open the tab inside the click, so it is not blocked as a popup,
            // then point it at the signed link once that arrives.
            const tab = window.open('about:blank', '_blank');
            void resolveUrl(href).then((link) => {
              if (tab) {
                tab.opener = null;
                tab.location.href = link;
              } else {
                window.open(link, '_blank', 'noopener,noreferrer');
              }
            });
          };
          return (
            <div
              key={a.id}
              className="flex items-center gap-2 px-2.5 py-2 rounded-lg group"
              style={{
                background: 'var(--flow-hover)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <button
                type="button"
                onClick={open}
                disabled={!href}
                className="flex flex-1 min-w-0 items-center gap-2.5 text-start disabled:cursor-default"
                aria-label={href ? t('flow.taskDetail.openAttachment', { name: a.name }) : a.name}
              >
                {isImage && href ? (
                  <AttachmentThumb url={href} name={a.name} />
                ) : (
                  <Icon
                    size={14}
                    className="shrink-0"
                    style={{ color: 'var(--flow-text-muted)' }}
                  />
                )}
                <span className="flex-1 min-w-0">
                  <span
                    className="block text-flow-meta font-medium truncate"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {a.name}
                  </span>
                  <span
                    className="block text-flow-meta"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {/* A name, not the raw user id. This read "47.0 KB · 9403",
                      crediting the uploader with a database key (rtylr#1079).
                      The attachment DTO carries only `uploadedBy`, so the id is
                      resolved against the member list here; an id with no
                      member still falls back to itself rather than to an empty
                      string, since a missing name should not read as nobody. */}
                    {formatSize(a.size)} · {uploaderName(a.uploadedBy)}
                  </span>
                </span>
              </button>
              {onRemove && !disabled && (canRemove?.(a) ?? true) && (
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
        className="flex items-center gap-2 px-3 py-2 rounded-lg text-flow-meta w-full transition-colors disabled:opacity-50"
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
        <p className="mt-1 text-flow-meta" style={{ color: 'var(--flow-danger-fg)' }}>
          {error}
        </p>
      )}
      <Dialog open={preview !== null} onOpenChange={(next) => !next && setPreview(null)}>
        <DialogContent className="max-w-[min(96vw,1100px)] p-3 sm:p-4">
          <DialogTitle className="pe-8 text-sm font-medium truncate">{preview?.name}</DialogTitle>
          <DialogDescription className="sr-only">
            {t('flow.taskDetail.attachmentPreview')}
          </DialogDescription>
          {preview && <PreviewBody url={preview.url} name={preview.name} />}
        </DialogContent>
      </Dialog>
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
