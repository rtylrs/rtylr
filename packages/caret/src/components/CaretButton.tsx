import React from 'react';
import { useCaretStore } from '../stores/caretStore';
import CaretLogo from './CaretLogo';
import { caretTokens } from '../ui/tokens';

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
        background: isOpen ? caretTokens.accentSoft : 'rgba(100, 197, 185, 0.1)',
        border: `1px solid ${isOpen ? caretTokens.accentBorder : 'rgba(100, 197, 185, 0.22)'}`,
        color: caretTokens.accent,
        cursor: 'pointer',
        fontSize: '11px',
        fontWeight: 600,
        boxShadow: isOpen ? `0 0 0 2px rgba(100, 197, 185, 0.1)` : 'none',
        transition: 'background 0.15s, border-color 0.15s, box-shadow 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = caretTokens.accentSoft;
        (e.currentTarget as HTMLButtonElement).style.borderColor = caretTokens.accentBorder;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = isOpen
          ? caretTokens.accentSoft
          : 'rgba(100, 197, 185, 0.1)';
        (e.currentTarget as HTMLButtonElement).style.borderColor = isOpen
          ? caretTokens.accentBorder
          : 'rgba(100, 197, 185, 0.22)';
      }}
    >
      <CaretLogo variant="teal" size={15} />
      Caret AI
    </button>
  );
};

export default CaretButton;
