import React from 'react';
import { caretTokens, caretRadius } from './tokens';

export interface BadgeProps {
  tone?: 'neutral' | 'accent' | 'success' | 'warn' | 'danger';
  children: React.ReactNode;
}

const TONES: Record<NonNullable<BadgeProps['tone']>, { bg: string; fg: string }> = {
  neutral: { bg: caretTokens.surfaceHigh, fg: caretTokens.textMuted },
  accent: { bg: caretTokens.accentSoft, fg: caretTokens.accent },
  success: { bg: 'rgba(16,185,129,0.14)', fg: caretTokens.success },
  warn: { bg: 'rgba(245,158,11,0.14)', fg: caretTokens.warn },
  danger: { bg: 'rgba(239,68,68,0.14)', fg: caretTokens.danger },
};

export const Badge: React.FC<BadgeProps> = ({ tone = 'neutral', children }) => {
  const c = TONES[tone];
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        padding: '2px 6px',
        borderRadius: caretRadius.sm,
        background: c.bg,
        color: c.fg,
        fontSize: 10,
        fontWeight: 500,
        textTransform: 'uppercase',
        letterSpacing: 0.2,
      }}
    >
      {children}
    </span>
  );
};

export default Badge;
