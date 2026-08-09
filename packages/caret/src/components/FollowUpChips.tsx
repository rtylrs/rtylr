import React from 'react';
import type { CaretStarter } from '../lib/suggestions';
import { caretRadius, caretTokens } from '../ui/tokens';

export interface FollowUpChipsProps {
  items: CaretStarter[];
  onSelect?: (item: CaretStarter) => void;
}

export const FollowUpChips: React.FC<FollowUpChipsProps> = ({ items, onSelect }) => {
  if (items.length === 0) return null;

  return (
    <div
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: 6,
        marginTop: 10,
      }}
    >
      {items.map((item) => (
        <button
          key={item.prompt}
          className="caret-follow-up"
          type="button"
          onClick={() => onSelect?.(item)}
          style={{
            border: `1px solid ${caretTokens.border}`,
            background: 'transparent',
            color: caretTokens.textMuted,
            borderRadius: caretRadius.sm,
            padding: '6px 9px',
            fontSize: 12,
            lineHeight: '16px',
            cursor: 'pointer',
          }}
        >
          {item.label ?? item.prompt}
        </button>
      ))}
    </div>
  );
};

export default FollowUpChips;
