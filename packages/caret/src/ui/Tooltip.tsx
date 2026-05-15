import React, { useState } from 'react';
import { caretTokens, caretRadius } from './tokens';

export interface TooltipProps {
  label: string;
  side?: 'top' | 'bottom';
  children: React.ReactElement;
}

export const Tooltip: React.FC<TooltipProps> = ({ label, side = 'top', children }) => {
  const [show, setShow] = useState(false);
  return (
    <span
      style={{ position: 'relative', display: 'inline-flex' }}
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
      onFocus={() => setShow(true)}
      onBlur={() => setShow(false)}
    >
      {children}
      {show && (
        <span
          role="tooltip"
          style={{
            position: 'absolute',
            [side === 'top' ? 'bottom' : 'top']: 'calc(100% + 6px)',
            left: '50%',
            transform: 'translateX(-50%)',
            background: caretTokens.surfaceHigh,
            color: caretTokens.text,
            fontSize: 11,
            padding: '4px 6px',
            borderRadius: caretRadius.sm,
            border: `1px solid ${caretTokens.border}`,
            whiteSpace: 'nowrap',
            pointerEvents: 'none',
            zIndex: 100,
          }}
        >
          {label}
        </span>
      )}
    </span>
  );
};

export default Tooltip;
