import React from 'react';
import { caretTokens, caretRadius } from './tokens';

export interface PillProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean;
  icon?: React.ReactNode;
}

export const Pill: React.FC<React.PropsWithChildren<PillProps>> = ({
  active = false,
  icon,
  children,
  style,
  ...rest
}) => (
  <button
    type="button"
    {...rest}
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      padding: '4px 10px',
      borderRadius: caretRadius.xl,
      background: active ? caretTokens.accentSoft : 'transparent',
      color: active ? caretTokens.accent : caretTokens.textMuted,
      border: `1px solid ${active ? caretTokens.accentBorder : caretTokens.border}`,
      fontSize: 12,
      fontWeight: 500,
      cursor: 'pointer',
      transition: 'background 0.12s, color 0.12s, border-color 0.12s',
      whiteSpace: 'nowrap',
      ...style,
    }}
  >
    {icon}
    {children}
  </button>
);

export default Pill;
