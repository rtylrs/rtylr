import React from 'react';
import { CaretLogo, useCaretStore } from '@voxire/caret';

/**
 * Tiny "Ask Caret" trigger you can drop anywhere on the dash. Opens the
 * Caret side-panel when clicked.
 */
export const CaretTrigger: React.FC<{ label?: string }> = ({ label = 'Ask Caret' }) => {
  const open = useCaretStore((s) => s.open);
  return (
    <button
      type="button"
      onClick={() => open()}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '4px 10px',
        borderRadius: 8,
        background: 'rgba(100,197,185,0.14)',
        color: '#64c5b9',
        border: '1px solid rgba(100,197,185,0.4)',
        fontSize: 12,
        fontWeight: 500,
        cursor: 'pointer',
      }}
    >
      <CaretLogo variant="teal" size={13} scale={1.8} />
      {label}
    </button>
  );
};

export default CaretTrigger;
