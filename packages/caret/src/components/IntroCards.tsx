import React from 'react';
import { caretTokens, caretRadius } from '../ui/tokens';

export interface IntroCard {
  id: string;
  title: string;
  description: string;
  prompt: string;
}

export interface IntroCardsProps {
  cards: IntroCard[];
  onSelect: (prompt: string) => void;
}

/** A 2x2 grid of suggestion cards shown in the empty-state. */
export const IntroCards: React.FC<IntroCardsProps> = ({ cards, onSelect }) => (
  <div
    style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
      gap: 8,
      width: '100%',
      maxWidth: 360,
    }}
  >
    {cards.map((c) => (
      <button
        key={c.id}
        type="button"
        onClick={() => onSelect(c.prompt)}
        style={{
          textAlign: 'left',
          padding: '10px 12px',
          borderRadius: caretRadius.md,
          background: caretTokens.surface,
          border: `1px solid ${caretTokens.border}`,
          color: caretTokens.text,
          cursor: 'pointer',
        }}
      >
        <div style={{ fontSize: 12.5, fontWeight: 600, marginBottom: 4 }}>{c.title}</div>
        <div
          style={{
            fontSize: 11.5,
            color: caretTokens.textMuted,
            lineHeight: 1.4,
          }}
        >
          {c.description}
        </div>
      </button>
    ))}
  </div>
);

export default IntroCards;
