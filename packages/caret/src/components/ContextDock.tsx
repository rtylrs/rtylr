import React from 'react';
import { ArrowRight, CircleDot } from 'lucide-react';
import type { CaretContextSignal, CaretStarter } from '../lib/suggestions';
import { caretRadius, caretTokens } from '../ui/tokens';

export interface ContextDockProps {
  signals: CaretContextSignal[];
  action?: CaretStarter;
  onAction?: (item: CaretStarter) => void;
}

const PRIORITY_LABELS = new Set(['Focus', 'Screen', 'Visible', 'Filter', 'Watch']);

function orderedSignals(signals: CaretContextSignal[]): CaretContextSignal[] {
  const priority = signals.filter((signal) => PRIORITY_LABELS.has(signal.label));
  const rest = signals.filter((signal) => !PRIORITY_LABELS.has(signal.label));
  return [...priority, ...rest].slice(0, 4);
}

export const ContextDock: React.FC<ContextDockProps> = ({ signals, action, onAction }) => {
  const visibleSignals = orderedSignals(signals);
  if (visibleSignals.length === 0 && !action) return null;

  return (
    <div
      style={{
        borderTop: `1px solid ${caretTokens.border}`,
        background: 'rgba(255, 255, 255, 0.018)',
        padding: '8px 12px',
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        minHeight: 42,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          minWidth: 0,
          flex: 1,
          overflow: 'hidden',
        }}
      >
        <CircleDot size={13} style={{ color: caretTokens.textDim, flex: '0 0 auto' }} />
        {visibleSignals.map((signal) => (
          <span
            key={`${signal.label}:${signal.value}`}
            style={{
              minWidth: 0,
              display: 'inline-flex',
              alignItems: 'baseline',
              gap: 4,
              maxWidth: signal.label === 'Screen' || signal.label === 'Focus' ? 150 : 110,
              color: caretTokens.textMuted,
              fontSize: 11.5,
              lineHeight: '15px',
              whiteSpace: 'nowrap',
            }}
          >
            <span
              style={{
                color: signal.tone === 'warn' ? caretTokens.warn : caretTokens.textDim,
                fontSize: 10,
                fontWeight: 700,
                textTransform: 'uppercase',
              }}
            >
              {signal.label}
            </span>
            <span
              style={{
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                minWidth: 0,
              }}
            >
              {signal.value}
            </span>
          </span>
        ))}
      </div>

      {action && (
        <button
          type="button"
          onClick={() => onAction?.(action)}
          style={{
            flex: '0 0 auto',
            display: 'inline-flex',
            alignItems: 'center',
            gap: 5,
            maxWidth: 150,
            border: `1px solid ${caretTokens.border}`,
            background: caretTokens.surface,
            color: caretTokens.textMuted,
            borderRadius: caretRadius.sm,
            padding: '5px 7px',
            fontSize: 11.5,
            lineHeight: '15px',
            cursor: 'pointer',
          }}
        >
          <span
            style={{
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {action.label ?? 'Brief this page'}
          </span>
          <ArrowRight size={12} style={{ flex: '0 0 auto', color: caretTokens.textDim }} />
        </button>
      )}
    </div>
  );
};

export default ContextDock;
