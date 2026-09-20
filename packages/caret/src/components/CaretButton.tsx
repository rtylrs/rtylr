import React, { useState } from 'react';
import { useCaretStore } from '../stores/caretStore';
import CaretLogo from './CaretLogo';
import { caretTokens } from '../ui/tokens';

interface CaretButtonProps {
  className?: string;
  style?: React.CSSProperties;
  'data-tour'?: string;
}

export function caretButtonStyle(
  isOpen: boolean,
  hovered: boolean,
  style?: React.CSSProperties,
): React.CSSProperties {
  return {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    height: '28px',
    padding: '0 12px',
    borderRadius: '8px',
    backgroundColor: isOpen ? caretTokens.accentSoft : hovered ? caretTokens.hover : 'transparent',
    border: `1px solid ${caretTokens.accentBorder}`,
    color: caretTokens.accent,
    cursor: 'pointer',
    fontSize: '11px',
    fontWeight: 600,
    boxShadow: 'none',
    transition: 'background-color 0.15s, border-color 0.15s',
    ...style,
  };
}

const CaretButton = ({
  className,
  style,
  'data-tour': dataTour,
}: CaretButtonProps): React.JSX.Element => {
  const toggle = useCaretStore((state) => state.toggle);
  const isOpen = useCaretStore((state) => state.isOpen);
  const [hovered, setHovered] = useState(false);

  return (
    <button
      onClick={toggle}
      aria-label="Open Caret AI"
      aria-expanded={isOpen}
      data-tour={dataTour}
      className={className}
      style={caretButtonStyle(isOpen, hovered, style)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <CaretLogo variant="teal" size={21} scale={1.18} />
      Caret AI
    </button>
  );
};

export default CaretButton;
