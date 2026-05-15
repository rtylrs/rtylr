import React, { useState } from 'react';
import { Wrench, ChevronDown, ChevronRight, Check, X, AlertCircle } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import Badge from '../../ui/Badge';
import Spinner from '../../ui/Spinner';
import type { ToolCallBlock } from '../../types/message';

export interface ToolCallCardProps {
  call: ToolCallBlock;
  onApprove?: (id: string) => void;
  onDeny?: (id: string) => void;
}

const STATUS_TONE: Record<ToolCallBlock['status'], React.ComponentProps<typeof Badge>['tone']> = {
  pending: 'warn',
  approved: 'accent',
  denied: 'danger',
  running: 'accent',
  success: 'success',
  error: 'danger',
};

const STATUS_LABEL: Record<ToolCallBlock['status'], string> = {
  pending: 'Awaiting approval',
  approved: 'Approved',
  denied: 'Denied',
  running: 'Running',
  success: 'Success',
  error: 'Error',
};

export const ToolCallCard: React.FC<ToolCallCardProps> = ({ call, onApprove, onDeny }) => {
  const [open, setOpen] = useState(call.status === 'error' || call.status === 'pending');

  return (
    <div
      style={{
        background: caretTokens.surface,
        border: `1px solid ${caretTokens.border}`,
        borderRadius: caretRadius.md,
        margin: '4px 0',
        overflow: 'hidden',
      }}
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          width: '100%',
          padding: '8px 10px',
          background: 'transparent',
          border: 'none',
          color: caretTokens.text,
          fontSize: 12,
          cursor: 'pointer',
        }}
      >
        {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        <Wrench size={12} style={{ color: caretTokens.accent }} />
        <span style={{ flex: 1, textAlign: 'left' }}>
          <span style={{ fontWeight: 600 }}>{call.name}</span>
        </span>
        <Badge tone={STATUS_TONE[call.status]}>
          {call.status === 'running' && <Spinner size={9} />}
          {STATUS_LABEL[call.status]}
        </Badge>
      </button>

      {open && (
        <div
          style={{
            padding: '8px 10px',
            borderTop: `1px solid ${caretTokens.border}`,
            fontFamily: "SFMono-Regular, Menlo, Consolas, monospace",
            fontSize: 11,
            color: caretTokens.textMuted,
          }}
        >
          <div style={{ marginBottom: 6 }}>
            <div style={{ color: caretTokens.textDim, marginBottom: 2 }}>input</div>
            <pre
              style={{
                margin: 0,
                padding: 6,
                background: caretTokens.bg,
                borderRadius: caretRadius.sm,
                overflowX: 'auto',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                color: caretTokens.text,
              }}
            >
              {JSON.stringify(call.input ?? {}, null, 2)}
            </pre>
          </div>

          {call.output !== undefined && (
            <div style={{ marginBottom: 6 }}>
              <div style={{ color: caretTokens.textDim, marginBottom: 2 }}>output</div>
              <pre
                style={{
                  margin: 0,
                  padding: 6,
                  background: caretTokens.bg,
                  borderRadius: caretRadius.sm,
                  overflowX: 'auto',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: caretTokens.text,
                }}
              >
                {typeof call.output === 'string'
                  ? call.output
                  : JSON.stringify(call.output, null, 2)}
              </pre>
            </div>
          )}

          {call.error && (
            <div
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: 6,
                padding: 6,
                background: 'rgba(239,68,68,0.08)',
                color: caretTokens.danger,
                borderRadius: caretRadius.sm,
                fontFamily: 'inherit',
              }}
            >
              <AlertCircle size={12} style={{ marginTop: 1 }} />
              <span>{call.error}</span>
            </div>
          )}

          {call.status === 'pending' && (onApprove || onDeny) && (
            <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
              {onApprove && (
                <button
                  type="button"
                  onClick={() => onApprove(call.id)}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 4,
                    padding: '4px 10px',
                    borderRadius: caretRadius.sm,
                    background: caretTokens.accent,
                    color: '#fff',
                    border: 'none',
                    fontSize: 12,
                    cursor: 'pointer',
                  }}
                >
                  <Check size={12} /> Approve
                </button>
              )}
              {onDeny && (
                <button
                  type="button"
                  onClick={() => onDeny(call.id)}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 4,
                    padding: '4px 10px',
                    borderRadius: caretRadius.sm,
                    background: 'transparent',
                    color: caretTokens.textMuted,
                    border: `1px solid ${caretTokens.border}`,
                    fontSize: 12,
                    cursor: 'pointer',
                  }}
                >
                  <X size={12} /> Deny
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ToolCallCard;
