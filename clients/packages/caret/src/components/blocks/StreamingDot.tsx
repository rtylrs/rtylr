import React from 'react';
import { caretTokens } from '../../ui/tokens';

/** A blinking dot that signals the assistant is still streaming. */
export const StreamingDot: React.FC = () => (
  <span
    aria-hidden="true"
    style={{
      display: 'inline-block',
      width: 7,
      height: 14,
      marginLeft: 2,
      verticalAlign: 'text-bottom',
      background: caretTokens.accent,
      borderRadius: 2,
      animation: 'caret-blink 1s steps(1) infinite',
    }}
  />
);

export default StreamingDot;
