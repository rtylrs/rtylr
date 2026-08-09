import React, { useRef } from 'react';
import { AlertTriangle } from 'lucide-react';
import { caretTokens, caretRadius } from '../ui/tokens';
import { useFocusTrap } from '../hooks/useFocusTrap';

export interface ApprovalDialogProps {
  open: boolean;
  toolName: string;
  description?: string;
  input: Record<string, unknown>;
  onApprove: () => void;
  onDeny: () => void;
}

/** Inline confirmation popover anchored above the composer. */
export const ApprovalDialog: React.FC<ApprovalDialogProps> = ({
  open,
  toolName,
  description,
  input,
  onApprove,
  onDeny,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  useFocusTrap(ref, open);

  if (!open) return null;

  return (
    <div
      ref={ref}
      role="dialog"
      aria-modal="true"
      aria-label={`Approve ${toolName}`}
      style={{
        margin: '0 12px',
        background: caretTokens.warnSoft,
        border: `1px solid ${caretTokens.warn}`,
        borderRadius: caretRadius.md,
        padding: 10,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: 8,
          marginBottom: 8,
        }}
      >
        <AlertTriangle size={14} style={{ color: caretTokens.warn, marginTop: 1 }} />
        <div style={{ flex: 1 }}>
          <div
            style={{
              color: caretTokens.text,
              fontSize: 13,
              fontWeight: 600,
              marginBottom: 2,
            }}
          >
            Caret wants to run <code style={{ fontFamily: 'inherit' }}>{toolName}</code>
          </div>
          {description && (
            <div style={{ color: caretTokens.textMuted, fontSize: 12 }}>{description}</div>
          )}
          <pre
            style={{
              margin: '6px 0 0',
              padding: 6,
              background: caretTokens.surfaceHigh,
              borderRadius: caretRadius.sm,
              fontSize: 11,
              color: caretTokens.textMuted,
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              maxHeight: 100,
              overflowY: 'auto',
            }}
          >
            {JSON.stringify(input, null, 2)}
          </pre>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
        <button
          type="button"
          onClick={onDeny}
          style={{
            padding: '5px 10px',
            background: 'transparent',
            color: caretTokens.textMuted,
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.sm,
            fontSize: 12,
            cursor: 'pointer',
          }}
        >
          Deny
        </button>
        <button
          type="button"
          onClick={onApprove}
          style={{
            padding: '5px 10px',
            background: caretTokens.accent,
            color: caretTokens.accentInk,
            border: 'none',
            borderRadius: caretRadius.sm,
            fontSize: 12,
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          Approve
        </button>
      </div>
    </div>
  );
};

export default ApprovalDialog;
