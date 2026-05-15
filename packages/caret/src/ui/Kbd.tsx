import React from 'react';
import { caretTokens, caretRadius } from './tokens';

export interface KbdProps {
  children: React.ReactNode;
}

export const Kbd: React.FC<KbdProps> = ({ children }) => (
  <kbd
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      padding: '0 5px',
      height: 18,
      borderRadius: caretRadius.sm,
      background: caretTokens.surfaceHigh,
      color: caretTokens.textMuted,
      fontFamily: 'inherit',
      fontSize: 11,
      lineHeight: 1,
      border: `1px solid ${caretTokens.border}`,
    }}
  >
    {children}
  </kbd>
);

export default Kbd;
