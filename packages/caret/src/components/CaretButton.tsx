import React from 'react';
import { useCaretStore } from '../stores/caretStore';
import CaretLogo from './CaretLogo';
import { caretTokens } from '../ui/tokens';

interface CaretButtonProps {
  className?: string;
  'data-tour'?: string;
}

const CaretButton = ({ className, 'data-tour': dataTour }: CaretButtonProps): React.JSX.Element => {
  const toggle = useCaretStore((state) => state.toggle);

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
        minHeight: '44px',
        minWidth: '44px',
        padding: '0 12px',
        borderRadius: '4px',
        background: caretTokens.accentSoft,
        border: `1px solid ${caretTokens.accentBorder}`,
        color: caretTokens.accent,
        cursor: 'pointer',
        fontSize: '11px',
        fontWeight: 600,
        boxShadow: 'none',
        transition: 'background 0.15s, border-color 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = caretTokens.accentSoft;
        (e.currentTarget as HTMLButtonElement).style.borderColor = caretTokens.accentBorder;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLButtonElement).style.background = caretTokens.accentSoft;
        (e.currentTarget as HTMLButtonElement).style.borderColor = caretTokens.accentBorder;
      }}
    >
      <CaretLogo variant="teal" size={21} scale={1.18} />
      Caret AI
    </button>
  );
};

export default CaretButton;
