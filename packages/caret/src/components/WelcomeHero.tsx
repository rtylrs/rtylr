import React from 'react';
import { caretTokens, caretRadius } from '../ui/tokens';
import { CARET_MODE_ICONS } from '../lib/modeIcons';
import CaretLogo from './CaretLogo';
import type { CaretEmptyStateModel, CaretStarter } from '../lib/suggestions';

export interface WelcomeHeroProps {
  emptyState: CaretEmptyStateModel;
  onStarter: (starter: CaretStarter) => void;
  /** 'lg' for the full-screen/overlay home experience, 'sm' for the inline side panel. */
  size?: 'sm' | 'lg';
}

function timeGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 5) return 'Still up.';
  if (hour < 12) return 'Good morning.';
  if (hour < 18) return 'Good afternoon.';
  return 'Good evening.';
}

/**
 * The Caret welcome screen - shown whenever a conversation has no messages yet.
 * A short greeting and a single tidy list of suggestions; the composer is
 * docked separately by the panel so it stays put whether this is showing or not.
 */
export const WelcomeHero: React.FC<WelcomeHeroProps> = ({ emptyState, onStarter, size = 'sm' }) => {
  const isLarge = size === 'lg';
  const rows = [emptyState.primaryAction, ...emptyState.suggestions].slice(0, isLarge ? 5 : 4);

  return (
    <div
      className="caret-empty-state"
      style={{
        flex: 1,
        minHeight: 0,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        overflowY: 'auto',
        padding: isLarge ? '48px 32px' : '20px 16px',
        gap: isLarge ? 24 : 14,
        textAlign: 'center',
      }}
    >
      <span className="caret-visually-hidden">
        {emptyState.title}. {emptyState.composerPlaceholder}
      </span>

      <div className="caret-hero-rise" aria-hidden="true">
        {isLarge && (
          <div style={{ marginBottom: 10 }}>
            <CaretLogo variant="teal" size={36} />
          </div>
        )}
        <div
          style={{
            fontSize: isLarge ? 22 : 14,
            fontWeight: 700,
            letterSpacing: '-0.01em',
            color: caretTokens.text,
          }}
        >
          {isLarge ? timeGreeting() : emptyState.title}
        </div>
        {isLarge && (
          <div style={{ fontSize: 14, color: caretTokens.textMuted, marginTop: 4 }}>
            {emptyState.title}
          </div>
        )}
      </div>

      {rows.length > 0 && (
        <div
          className="caret-hero-rise"
          style={{
            width: '100%',
            maxWidth: isLarge ? 520 : '100%',
            borderRadius: 14,
            border: `1px solid ${caretTokens.border}`,
            background: caretTokens.surface,
            overflow: 'hidden',
            textAlign: 'left',
            animationDelay: '60ms',
          }}
        >
          {rows.map((starter, index) => (
            <button
              key={`${starter.prompt}-${index}`}
              type="button"
              className="caret-hero-row"
              onClick={() => onStarter(starter)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                width: '100%',
                padding: isLarge ? '12px 16px' : '9px 11px',
                border: 'none',
                borderTop: index === 0 ? 'none' : `1px solid ${caretTokens.border}`,
                background: 'transparent',
                color: caretTokens.text,
                cursor: 'pointer',
                textAlign: 'left',
              }}
            >
              <span
                aria-hidden="true"
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: isLarge ? 32 : 26,
                  height: isLarge ? 32 : 26,
                  borderRadius: caretRadius.lg,
                  background: caretTokens.surfaceHigh,
                  color: caretTokens.accent,
                  flexShrink: 0,
                }}
              >
                {starter.mode ? CARET_MODE_ICONS[starter.mode] : null}
              </span>
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{ fontSize: isLarge ? 13.5 : 12.5, fontWeight: 600 }}>
                  {starter.label ?? starter.prompt}
                </div>
                {starter.description && (
                  <div
                    style={{
                      fontSize: isLarge ? 12 : 11.5,
                      color: caretTokens.textMuted,
                      marginTop: 1,
                    }}
                  >
                    {starter.description}
                  </div>
                )}
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default WelcomeHero;
