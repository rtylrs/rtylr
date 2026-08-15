import React, { useState } from 'react';
import { Search, ChevronDown, ChevronRight, Check, X, AlertCircle } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import Spinner from '../../ui/Spinner';
import type { ToolCallBlock } from '../../types/message';

export interface ToolCallCardProps {
  call: ToolCallBlock;
  onApprove?: (id: string) => void;
  onDeny?: (id: string) => void;
}

const detailLabelStyle: React.CSSProperties = {
  color: caretTokens.textDim,
  fontSize: 11,
  marginBottom: 4,
};

const summaryStyle: React.CSSProperties = {
  padding: '7px 9px',
  background: caretTokens.surfaceHigh,
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  color: caretTokens.textMuted,
  fontFamily: 'inherit',
  lineHeight: 1.45,
};

function actionForTool(name: string): string {
  const normalized = name.trim().toLowerCase();
  const actions: Array<[RegExp, string]> = [
    [/^erp\.create_product$/, 'add a product'],
    [/^erp\.update_product$/, 'update a product'],
    [/^erp\.delete_product$/, 'delete a product'],
    [/^erp\.set_stock$/, 'adjust product stock'],
    [/^erp\.list_reorder_suggestions$/, 'review reorder suggestions'],
    [/^.*\.create_/, 'create a business record'],
    [/^.*\.delete_/, 'delete a business record'],
    [/^.*\.update_/, 'update a business record'],
  ];
  return actions.find(([pattern]) => pattern.test(normalized))?.[1] ?? 'check business data';
}

function statusText(status: ToolCallBlock['status'], action: string): string {
  switch (status) {
    case 'pending':
      return `Needs confirmation to ${action}`;
    case 'approved':
    case 'running':
      return `Working to ${action}`;
    case 'success':
      return `Completed: ${action}`;
    case 'denied':
      return `Skipped: ${action}`;
    case 'error':
      return `Could not ${action}`;
  }
}

export const ToolCallCard: React.FC<ToolCallCardProps> = ({ call, onApprove, onDeny }) => {
  const [open, setOpen] = useState(call.status === 'pending');
  const action = actionForTool(call.name);
  const label = statusText(call.status, action);
  const isPending = call.status === 'pending';
  const isRunning = call.status === 'running' || call.status === 'approved';
  const quiet = !isPending && call.status !== 'error';
  const ActivityIcon = call.status === 'error' ? AlertCircle : Search;

  return (
    <div
      style={{
        margin: isPending ? '8px 0 10px' : '4px 0',
        color: caretTokens.textDim,
      }}
    >
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: quiet ? 7 : 9,
          maxWidth: '100%',
          padding: quiet ? '1px 0' : '2px 0',
          background: 'transparent',
          border: 'none',
          color: isPending ? caretTokens.textMuted : caretTokens.textDim,
          fontSize: quiet ? 12 : 13,
          lineHeight: 1.4,
          cursor: 'pointer',
          opacity: isPending ? 0.95 : quiet ? 0.54 : 0.72,
        }}
      >
        {open ? <ChevronDown size={quiet ? 12 : 13} /> : <ChevronRight size={quiet ? 12 : 13} />}
        {isRunning ? (
          <Spinner size={quiet ? 10 : 12} />
        ) : (
          <ActivityIcon size={quiet ? 12 : 14} style={{ flex: '0 0 auto' }} />
        )}
        <span style={{ minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis' }}>{label}</span>
      </button>

      {open && (
        <div
          style={{
            marginTop: quiet ? 4 : 6,
            marginLeft: quiet ? 18 : 21,
            padding: quiet ? '3px 0 1px 10px' : '4px 0 2px 12px',
            borderLeft: `1px solid ${caretTokens.border}`,
            fontFamily: 'inherit',
            fontSize: 11,
            color: caretTokens.textMuted,
            opacity: quiet ? 0.88 : 1,
          }}
        >
          {call.preview && (
            <div style={{ marginBottom: 8 }}>
              <div style={detailLabelStyle}>what will happen</div>
              <div style={summaryStyle}>{call.preview}</div>
            </div>
          )}

          {!call.preview && (
            <div style={{ ...summaryStyle, marginBottom: 8 }}>
              Caret will {action}. Review the request and confirm only if that is what you want.
            </div>
          )}

          {call.summary && (
            <div style={{ marginBottom: 8 }}>
              <div style={detailLabelStyle}>summary</div>
              <div style={summaryStyle}>{call.summary}</div>
            </div>
          )}

          {call.error && (
            <div
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: 6,
                padding: '8px 9px',
                background: caretTokens.surfaceHigh,
                border: `1px solid ${caretTokens.border}`,
                color: caretTokens.textMuted,
                borderRadius: caretRadius.sm,
                fontFamily: 'inherit',
              }}
            >
              <AlertCircle size={12} style={{ marginTop: 1, color: caretTokens.textDim }} />
              <span>{call.error}</span>
            </div>
          )}

          {call.status === 'pending' && (onApprove || onDeny) && (
            <div style={{ display: 'flex', gap: 6, marginTop: 9, fontFamily: 'inherit' }}>
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
                    background: caretTokens.accentSoft,
                    color: caretTokens.text,
                    border: `1px solid ${caretTokens.accentBorder}`,
                    fontSize: 12,
                    cursor: 'pointer',
                  }}
                >
                  <Check size={12} /> Allow
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
                  <X size={12} /> Skip
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
