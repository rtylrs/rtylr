import React from 'react';
import { Sparkles } from 'lucide-react';
import { caretTokens, caretRadius } from '../ui/tokens';

export interface EmptyStateProps {
  title?: string;
  subtitle?: string;
  suggestions?: string[];
  onSuggestion?: (s: string) => void;
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
          {suggestions.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => onSuggestion?.(s)}
              style={{
                padding: '8px 10px',
                background: caretTokens.surface,
                border: `1px solid ${caretTokens.border}`,
                borderRadius: caretRadius.md,
                color: caretTokens.text,
                fontSize: 13,
                textAlign: 'left',
                cursor: 'pointer',
              }}
            >
              {s}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default EmptyState;
