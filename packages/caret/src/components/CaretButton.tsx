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
        background: caretTokens.accent,
        border: `1px solid ${caretTokens.accentBorder}`,
        color: '#ffffff',
        cursor: 'pointer',
        fontSize: '11px',
        fontWeight: 600,
        boxShadow: isOpen ? `0 0 0 3px ${caretTokens.accentSoft}` : 'none',
        transition: 'filter 0.15s, box-shadow 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLButtonElement).style.filter = 'brightness(1.08)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLButtonElement).style.filter = 'none';
      }}
    >
      <CaretLogo variant="white" size={15} scale={1.85} />
      Caret AI
    </button>
  );
};

export default CaretButton;
