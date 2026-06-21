import React from 'react';
import { ArrowRight } from 'lucide-react';
import type { CaretContextSignal, CaretStarter } from '../lib/suggestions';
import { caretRadius, caretTokens } from '../ui/tokens';

export interface ContextDockProps {
  signals: CaretContextSignal[];
  action?: CaretStarter;
  onAction?: (item: CaretStarter) => void;
}

export const ContextDock: React.FC<ContextDockProps> = ({ action, onAction }) => {
  if (!action) return null;

  return (
    <div
      style={{
        borderTop: `1px solid ${caretTokens.border}`,
        background: 'transparent',
        padding: '7px 12px',
        display: 'flex',
        justifyContent: 'flex-end',
        minHeight: 36,
      }}
    >
      <button
        type="button"
        onClick={() => onAction?.(action)}
        style={{
          flex: '0 0 auto',
          display: 'inline-flex',
          alignItems: 'center',
          gap: 5,
          maxWidth: 160,
          border: `1px solid ${caretTokens.border}`,
          background: 'transparent',
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
          {action.label ?? 'Ask about this'}
        </span>
        <ArrowRight size={12} style={{ flex: '0 0 auto', color: caretTokens.textDim }} />
      </button>
    </div>
  );
};

export default ContextDock;
