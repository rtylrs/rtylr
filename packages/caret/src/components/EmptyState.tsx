import React from 'react';
import type { CaretContextSignal, CaretStarter } from '../lib/suggestions';
import { caretTokens, caretRadius } from '../ui/tokens';
import CaretLogo from './CaretLogo';

export interface EmptyStateProps {
  title?: string;
  subtitle?: string;
  briefItems?: string[];
  contextSignals?: CaretContextSignal[];
  primaryAction?: CaretStarter;
  suggestions?: CaretStarter[];
  onSuggestion?: (s: CaretStarter) => void;
}

function compactStarters(
  primaryAction: CaretStarter | undefined,
  suggestions: CaretStarter[]
): CaretStarter[] {
  const seen = new Set<string>();
  const result: CaretStarter[] = [];

  for (const item of [primaryAction, ...suggestions]) {
    if (!item?.prompt) continue;
    const key = item.prompt.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(item);
    if (result.length >= 4) break;
  }

  return result;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'Ask Caret',
  subtitle,
  primaryAction,
  suggestions = [],
  onSuggestion,
}) => {
  const visibleSuggestions = compactStarters(primaryAction, suggestions);

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px 18px',
        overflowY: 'auto',
      }}
    >
      <section
        style={{
          width: '100%',
          maxWidth: 340,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: 12,
          }}
        >
          <CaretLogo variant="teal" size={30} />
        </div>
        <p
          style={{
            color: caretTokens.text,
            fontSize: 15.5,
            fontWeight: 700,
            margin: subtitle || visibleSuggestions.length > 0 ? '0 0 10px' : 0,
            textAlign: 'center',
          }}
        >
          {title}
        </p>
        {subtitle ? (
          <p
            style={{
              color: caretTokens.textMuted,
              fontSize: 12.5,
              lineHeight: 1.45,
              textAlign: 'center',
              margin: visibleSuggestions.length > 0 ? '0 0 14px' : 0,
              maxWidth: 250,
            }}
          >
            {subtitle}
          </p>
        ) : null}

        {visibleSuggestions.length > 0 && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(138px, 1fr))',
              gap: 6,
              width: '100%',
            }}
          >
            {visibleSuggestions.map((s) => (
              <button
                key={s.prompt}
                type="button"
                onClick={() => onSuggestion?.(s)}
                style={{
                  minHeight: 40,
                  padding: '8px 9px',
                  background: caretTokens.surface,
                  border: `1px solid ${caretTokens.border}`,
                  borderRadius: caretRadius.md,
                  color: caretTokens.text,
                  cursor: 'pointer',
                  textAlign: 'left',
                }}
              >
                <span
                  style={{
                    display: 'block',
                    fontSize: 12.5,
                    fontWeight: 600,
                    lineHeight: 1.2,
                  }}
                >
                  {s.label ?? s.prompt}
                </span>
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
};

export default EmptyState;
