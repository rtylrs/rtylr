import React from 'react';
import { ArrowRight, Sparkles } from 'lucide-react';
import type { CaretContextSignal, CaretStarter } from '../lib/suggestions';
import { caretTokens, caretRadius } from '../ui/tokens';

export interface EmptyStateProps {
  title?: string;
  subtitle?: string;
  contextSignals?: CaretContextSignal[];
  primaryAction?: CaretStarter;
  suggestions?: CaretStarter[];
  onSuggestion?: (s: CaretStarter) => void;
}

function signalTone(signal: CaretContextSignal): React.CSSProperties {
  if (signal.tone === 'accent') {
    return {
      border: `1px solid ${caretTokens.accentBorder}`,
      background: caretTokens.accentSoft,
      color: caretTokens.text,
    };
  }
  if (signal.tone === 'warn') {
    return {
      border: '1px solid rgba(245, 158, 11, 0.35)',
      background: 'rgba(245, 158, 11, 0.12)',
      color: caretTokens.text,
    };
  }
  return {
    border: `1px solid ${caretTokens.border}`,
    background: caretTokens.surface,
    color: caretTokens.textMuted,
  };
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'Ask Caret anything',
  subtitle = 'Your AI assistant with access to all your business data.',
  contextSignals = [],
  primaryAction,
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
        overflowY: 'auto',
      }}
    >
      <section
        style={{
          width: '100%',
          maxWidth: 380,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
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
            fontSize: 15,
            fontWeight: 700,
            margin: '0 0 4px',
            textAlign: 'center',
          }}
        >
          {title}
        </p>
        <p
          style={{
            color: caretTokens.textMuted,
            fontSize: 13,
            lineHeight: 1.45,
            textAlign: 'center',
            margin: '0 0 16px',
            maxWidth: 310,
          }}
        >
          {subtitle}
        </p>

        {contextSignals.length > 0 && (
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              gap: 6,
              width: '100%',
              marginBottom: 14,
            }}
          >
            {contextSignals.map((signal) => (
              <span
                key={`${signal.label}:${signal.value}`}
                style={{
                  ...signalTone(signal),
                  display: 'inline-flex',
                  alignItems: 'baseline',
                  gap: 5,
                  maxWidth: '100%',
                  borderRadius: caretRadius.sm,
                  padding: '5px 7px',
                  fontSize: 11.5,
                  lineHeight: '15px',
                  whiteSpace: 'nowrap',
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    color: caretTokens.textDim,
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
        )}

        {primaryAction && (
          <button
            type="button"
            onClick={() => onSuggestion?.(primaryAction)}
            style={{
              width: '100%',
              minHeight: 68,
              padding: '12px 13px',
              background: caretTokens.accentSoft,
              border: `1px solid ${caretTokens.accentBorder}`,
              borderRadius: caretRadius.md,
              color: caretTokens.text,
              cursor: 'pointer',
              textAlign: 'left',
              marginBottom: suggestions.length > 0 ? 8 : 0,
            }}
          >
            <span
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 10,
                fontSize: 13.5,
                fontWeight: 700,
                lineHeight: 1.25,
              }}
            >
              {primaryAction.label ?? primaryAction.prompt}
              <ArrowRight size={15} style={{ flex: '0 0 auto', color: caretTokens.accent }} />
            </span>
            {primaryAction.description && (
              <span
                style={{
                  display: 'block',
                  color: caretTokens.textMuted,
                  fontSize: 12,
                  lineHeight: 1.4,
                  marginTop: 5,
                }}
              >
                {primaryAction.description}
              </span>
            )}
          </button>
        )}

        {suggestions.length > 0 && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(148px, 1fr))',
              gap: 6,
              width: '100%',
            }}
          >
            {suggestions.map((s) => (
              <button
                key={s.prompt}
                type="button"
                onClick={() => onSuggestion?.(s)}
                style={{
                  minHeight: 58,
                  padding: '9px 10px',
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
                    lineHeight: 1.25,
                  }}
                >
                  {s.label ?? s.prompt}
                </span>
                {s.description && (
                  <span
                    style={{
                      display: 'block',
                      color: caretTokens.textMuted,
                      fontSize: 11.5,
                      lineHeight: 1.35,
                      marginTop: 4,
                    }}
                  >
                    {s.description}
                  </span>
                )}
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
};

export default EmptyState;
