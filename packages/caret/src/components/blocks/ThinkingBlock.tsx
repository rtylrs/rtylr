import React, { useState } from 'react';
import { Brain, ChevronDown, ChevronRight } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';

export interface ThinkingBlockProps {
  content: string;
  durationMs?: number;
  /** Whether the model is still thinking. Shows a pulsing dot when true. */
  isLive?: boolean;
  defaultOpen?: boolean;
}

export const ThinkingBlock: React.FC<ThinkingBlockProps> = ({
  content,
  durationMs,
  isLive = false,
  defaultOpen = false,
}) => {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div
      style={{
        background: caretTokens.accentSoft,
        border: `1px solid ${caretTokens.accentBorder}`,
        borderRadius: caretRadius.md,
        padding: '6px 8px',
        margin: '4px 0',
      }}
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          width: '100%',
          background: 'transparent',
          border: 'none',
          color: caretTokens.thinking,
          fontSize: 12,
          fontWeight: 500,
          cursor: 'pointer',
        }}
      >
        {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        <Brain size={12} />
        <span>{isLive ? 'Thinking…' : 'Thought process'}</span>
        {isLive && (
          <span
            aria-hidden="true"
            style={{
              width: 6,
              height: 6,
              borderRadius: 3,
              background: caretTokens.thinking,
              animation: 'caret-pulse 1.2s ease-in-out infinite',
            }}
          />
        )}
        {!isLive && typeof durationMs === 'number' && (
          <span style={{ marginLeft: 'auto', color: caretTokens.textMuted, fontSize: 11 }}>
            {Math.max(1, Math.round(durationMs / 100) / 10)}s
          </span>
        )}
      </button>
      {open && (
        <div
          style={{
            marginTop: 6,
            padding: '6px 8px',
            background: caretTokens.surface,
            borderRadius: caretRadius.sm,
            color: caretTokens.textMuted,
            fontSize: 12,
            lineHeight: 1.55,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {content || (isLive ? '…' : '')}
        </div>
      )}
    </div>
  );
};

export default ThinkingBlock;
