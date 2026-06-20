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

const TOOL_TARGETS: Array<[RegExp, string]> = [
  [/^analytics\./, 'analytics'],
  [/^crm\./, 'customer data'],
  [/^customers?\./, 'customer data'],
  [/^inventory\./, 'inventory'],
  [/^orders?\./, 'orders'],
  [/^finance\./, 'finance'],
  [/^booking\./, 'bookings'],
  [/^(hr|workforce|employees?)\./, 'team data'],
  [/^loyalty\./, 'loyalty data'],
  [/^(menu|catalog)\./, 'menu data'],
  [/^(marketing|growth|insights)\./, 'business insights'],
  [/^(operator|operations)\./, 'operations'],
  [/^settings\./, 'settings'],
  [/^locations?\./, 'locations'],
  [/^payments?\./, 'payments'],
];

const detailLabelStyle: React.CSSProperties = {
  color: caretTokens.textDim,
  fontSize: 11,
  marginBottom: 4,
};

const detailPreStyle: React.CSSProperties = {
  margin: 0,
  padding: '8px 9px',
  background: 'rgba(255,255,255,0.03)',
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  overflowX: 'auto',
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
  color: caretTokens.textMuted,
};

const summaryStyle: React.CSSProperties = {
  padding: '7px 9px',
  background: 'rgba(255,255,255,0.025)',
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  color: caretTokens.textMuted,
  fontFamily: 'inherit',
  lineHeight: 1.45,
};

function targetForTool(name: string): string {
  const normalized = name.trim().toLowerCase();
  const match = TOOL_TARGETS.find(([pattern]) => pattern.test(normalized));

  if (match) return match[1];

  const [moduleName] = normalized.split('.');
  const readableModule = moduleName?.replace(/[_-]/g, ' ').trim();
  return readableModule ? `${readableModule} data` : 'business data';
}

function statusText(status: ToolCallBlock['status'], target: string): string {
  switch (status) {
    case 'pending':
      return `Needs permission to check ${target}`;
    case 'approved':
    case 'running':
      return `Checking ${target}`;
    case 'success':
      return `Checked ${target}`;
    case 'denied':
      return `Skipped checking ${target}`;
    case 'error':
      return `Could not check ${target}`;
  }
}

function formatValue(value: unknown): string {
  if (typeof value === 'string') return value;

  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return String(value);
  }
}

export const ToolCallCard: React.FC<ToolCallCardProps> = ({ call, onApprove, onDeny }) => {
  const [open, setOpen] = useState(call.status === 'pending');
  const target = targetForTool(call.name);
  const label = statusText(call.status, target);
  const isPending = call.status === 'pending';
  const isRunning = call.status === 'running' || call.status === 'approved';
  const ActivityIcon = call.status === 'error' ? AlertCircle : Search;

  return (
    <div
      style={{
        margin: isPending ? '8px 0 10px' : '8px 0',
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
          gap: 9,
          maxWidth: '100%',
          padding: '2px 0',
          background: 'transparent',
          border: 'none',
          color: isPending ? caretTokens.textMuted : caretTokens.textDim,
          fontSize: 13,
          lineHeight: 1.4,
          cursor: 'pointer',
          opacity: isPending ? 0.95 : 0.76,
        }}
      >
        {open ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
        {isRunning ? (
          <Spinner size={12} />
        ) : (
          <ActivityIcon size={14} style={{ flex: '0 0 auto' }} />
        )}
        <span style={{ minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis' }}>{label}</span>
      </button>

      {open && (
        <div
          style={{
            marginTop: 6,
            marginLeft: 21,
            padding: '4px 0 2px 12px',
            borderLeft: `1px solid ${caretTokens.border}`,
            fontFamily: 'SFMono-Regular, Menlo, Consolas, monospace',
            fontSize: 11,
            color: caretTokens.textMuted,
          }}
        >
          <div style={{ marginBottom: 8 }}>
            <div style={detailLabelStyle}>source</div>
            <pre style={detailPreStyle}>{call.name}</pre>
          </div>

          <div style={{ marginBottom: 8 }}>
            <div style={detailLabelStyle}>details</div>
            <pre style={detailPreStyle}>{formatValue(call.input ?? {})}</pre>
          </div>

          {call.summary && (
            <div style={{ marginBottom: 8 }}>
              <div style={detailLabelStyle}>summary</div>
              <div style={summaryStyle}>{call.summary}</div>
            </div>
          )}

          {call.output !== undefined && (
            <div style={{ marginBottom: 8 }}>
              <div style={detailLabelStyle}>result</div>
              <pre style={detailPreStyle}>{formatValue(call.output)}</pre>
            </div>
          )}

          {call.error && (
            <div
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: 6,
                padding: '8px 9px',
                background: 'rgba(255,255,255,0.03)',
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
