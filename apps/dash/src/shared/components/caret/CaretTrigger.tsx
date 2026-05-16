import React from 'react';
import { Sparkles } from 'lucide-react';
import { useCaretStore } from '@voxire/caret';

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
        background: 'rgba(139,92,246,0.14)',
        color: '#8b5cf6',
        border: '1px solid rgba(139,92,246,0.4)',
        fontSize: 12,
        fontWeight: 500,
        cursor: 'pointer',
      }}
    >
      <Sparkles size={12} />
      {label}
    </button>
  );
};

export default CaretTrigger;
