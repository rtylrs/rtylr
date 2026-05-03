import React from 'react';
import { Sparkles } from 'lucide-react';
import { useCaretStore } from '../stores/caretStore';

interface CaretButtonProps {
  className?: string;
  'data-tour'?: string;
}

const CaretButton = ({ className, 'data-tour': dataTour }: CaretButtonProps): React.JSX.Element => {
  const { isOpen, toggle } = useCaretStore();

  return (
    <button
      onClick={toggle}
      aria-label="Open Caret AI"
      data-tour={dataTour}
      className={className}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        height: '28px',
        padding: '0 12px',
        borderRadius: '8px',
        background: isOpen ? 'rgba(139,92,246,0.18)' : 'rgba(139,92,246,0.1)',
        border: `1px solid ${isOpen ? 'rgba(139,92,246,0.4)' : 'rgba(139,92,246,0.2)'}`,
        color: '#8b5cf6',
        cursor: 'pointer',
        fontSize: '11px',
        fontWeight: 600,
        transition: 'background 0.15s, border-color 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = 'rgba(139,92,246,0.18)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = isOpen
          ? 'rgba(139,92,246,0.18)'
          : 'rgba(139,92,246,0.1)';
      }}
    >
      <Sparkles size={13} aria-hidden="true" />
      Caret AI
    </button>
  );
};

export default CaretButton;
