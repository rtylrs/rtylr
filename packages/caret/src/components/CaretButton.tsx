import React from 'react';
import { Sparkles } from 'lucide-react';
import { useCaretStore } from '../stores/caretStore';

interface CaretButtonProps {
  className?: string;
}

const CaretButton = ({ className }: CaretButtonProps): React.JSX.Element => {
  const { isOpen, toggle } = useCaretStore();

  return (
    <button
      onClick={toggle}
      aria-label="Open Caret AI"
      className={className}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        height: '36px',
        padding: '0 14px',
        borderRadius: '18px',
        background: isOpen ? 'var(--caret-accent, #8b5cf6)' : 'transparent',
        border: `1px solid ${isOpen ? 'var(--caret-accent, #8b5cf6)' : 'var(--caret-border, #2a2a2a)'}`,
        color: isOpen ? '#ffffff' : 'var(--caret-text-muted, #888888)',
        cursor: 'pointer',
        fontSize: '13px',
        fontWeight: 500,
        transition: 'background 0.15s, border-color 0.15s, color 0.15s',
        position: 'relative',
      }}
    >
      <Sparkles
        size={14}
        style={{ color: isOpen ? '#ffffff' : 'var(--caret-accent, #8b5cf6)' }}
        aria-hidden="true"
      />
      Caret
      {isOpen && (
        <span
          aria-hidden="true"
          style={{
            position: 'absolute',
            top: '4px',
            right: '4px',
            width: '6px',
            height: '6px',
            borderRadius: '50%',
            background: '#ffffff',
          }}
        />
      )}
    </button>
  );
};

export default CaretButton;
