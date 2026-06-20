import React from 'react';
import { Sparkles } from 'lucide-react';
import type { CaretStarter } from '../lib/suggestions';
import { caretTokens, caretRadius } from '../ui/tokens';

export interface EmptyStateProps {
  title?: string;
  subtitle?: string;
  suggestions?: CaretStarter[];
  onSuggestion?: (s: CaretStarter) => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'Ask Caret anything',
  subtitle = 'Your AI assistant with access to all your business data.',
  suggestions = [],
  onSuggestion,
}) => {
  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px 16px',
      }}
    >
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: caretRadius.xl,
          background: caretTokens.accentSoft,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginBottom: 14,
        }}
      >
        <Sparkles size={22} style={{ color: caretTokens.accent }} />
      </div>
      <p
        style={{
          color: caretTokens.text,
          fontSize: 14,
          fontWeight: 600,
          margin: '0 0 4px',
        }}
      >
        {title}
      </p>
      <p
        style={{
          color: caretTokens.textMuted,
          fontSize: 13,
          textAlign: 'center',
          margin: '0 0 20px',
          maxWidth: 280,
        }}
      >
        {subtitle}
      </p>
      {suggestions.length > 0 && (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 6,
            width: '100%',
            maxWidth: 320,
          }}
        >
          {suggestions.map((s, index) => (
            <button
              key={s.prompt}
              type="button"
              onClick={() => onSuggestion?.(s)}
              style={{
                padding: index === 0 ? '10px 11px' : '8px 10px',
                background: index === 0 ? caretTokens.accentSoft : caretTokens.surface,
                border: `1px solid ${index === 0 ? caretTokens.accentBorder : caretTokens.border}`,
                borderRadius: caretRadius.md,
                color: caretTokens.text,
                fontSize: 13,
                fontWeight: index === 0 ? 600 : 500,
                textAlign: 'left',
                cursor: 'pointer',
              }}
            >
              {s.label ?? s.prompt}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default EmptyState;
