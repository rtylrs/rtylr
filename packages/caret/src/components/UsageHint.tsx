import React from 'react';
import { Info } from 'lucide-react';
import { caretTokens, caretRadius } from '../ui/tokens';

export interface UsageHintProps {
  children: React.ReactNode;
}

/** Subtle inline hint used by the composer (mode tips, token warnings). */
export const UsageHint: React.FC<UsageHintProps> = ({ children }) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'flex-start',
      gap: 6,
      padding: '4px 6px',
      borderRadius: caretRadius.sm,
      background: caretTokens.surfaceHigh,
      color: caretTokens.textMuted,
      fontSize: 11.5,
      lineHeight: 1.45,
    }}
  >
    <Info size={11} style={{ marginTop: 1, flexShrink: 0, color: caretTokens.textDim }} />
    <span>{children}</span>
  </div>
);

export default UsageHint;
