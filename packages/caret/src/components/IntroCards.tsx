import React from 'react';
import { caretTokens, caretRadius } from '../ui/tokens';
import { CARET_MODE_ICONS } from '../lib/modeIcons';
import { CARET_MODES } from '../types/mode';
import type { CaretStarter } from '../lib/suggestions';

export interface IntroCardsProps {
  starters: CaretStarter[];
  onSelect: (starter: CaretStarter) => void;
  size?: 'sm' | 'lg';
}

/** A responsive grid of suggestion cards shown on the Caret welcome screen. */
export const IntroCards: React.FC<IntroCardsProps> = ({ starters, onSelect, size = 'sm' }) => (
  <div
    className="caret-hero-grid"
    style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
      gap: size === 'lg' ? 14 : 7,
      width: '100%',
      maxWidth: size === 'lg' ? 760 : 340,
      margin: '0 auto',
    }}
  >
    {starters.map((starter, index) => {
      const modeDef = starter.mode ? CARET_MODES[starter.mode] : null;
      return (
        <button
          key={`${starter.prompt}-${index}`}
          type="button"
          className="caret-hero-card"
          onClick={() => onSelect(starter)}
          style={{
            textAlign: 'left',
            padding: size === 'lg' ? '20px 22px' : '10px 12px',
            borderRadius: size === 'lg' ? 14 : caretRadius.lg,
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            color: caretTokens.text,
            cursor: 'pointer',
            animationDelay: `${index * 35}ms`,
          }}
        >
          {modeDef && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                padding: size === 'lg' ? '3px 9px' : '2px 7px',
                marginBottom: size === 'lg' ? 10 : 6,
                borderRadius: caretRadius.xl,
                background: caretTokens.accentSoft,
                color: caretTokens.accent,
                fontSize: size === 'lg' ? 11 : 10,
                fontWeight: 600,
              }}
            >
              {CARET_MODE_ICONS[modeDef.id]}
              {modeDef.label}
            </span>
          )}
          <div
            style={{
              fontSize: size === 'lg' ? 15 : 12.5,
              fontWeight: 600,
              marginBottom: size === 'lg' ? 5 : 3,
            }}
          >
            {starter.label ?? starter.prompt}
          </div>
          {starter.description && (
            <div
              style={{
                fontSize: size === 'lg' ? 13 : 11.5,
                color: caretTokens.textMuted,
                lineHeight: 1.5,
              }}
            >
              {starter.description}
            </div>
          )}
        </button>
      );
    })}
  </div>
);

export default IntroCards;
