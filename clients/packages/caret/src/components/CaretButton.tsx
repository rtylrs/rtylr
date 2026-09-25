import React, { useState } from 'react';
import { useCaretStore } from '../stores/caretStore';
import CaretLogo from './CaretLogo';
import { caretTokens } from '../ui/tokens';

export type CaretButtonSize = 'sm' | 'md';

interface CaretButtonProps {
  className?: string;
  size?: CaretButtonSize;
  style?: React.CSSProperties;
  'data-tour'?: string;
}

const sizeStyles: Record<CaretButtonSize, React.CSSProperties> = {
  sm: {
    gap: '6px',
    height: '28px',
    padding: '0 12px',
    borderRadius: '8px',
    fontSize: '11px',
  },
  md: {
    gap: '8px',
    height: '36px',
    padding: '0 12px',
    borderRadius: '6px',
    fontSize: '12px',
  },
};

export function caretButtonStyle(
  size: CaretButtonSize,
  isOpen: boolean,
  hovered: boolean,
  style?: React.CSSProperties,
): React.CSSProperties {
  return {
    display: 'inline-flex',
    alignItems: 'center',
    ...sizeStyles[size],
    backgroundColor: isOpen ? caretTokens.accentSoft : hovered ? caretTokens.hover : 'transparent',
    border: `1px solid ${caretTokens.accentBorder}`,
    color: caretTokens.accent,
    cursor: 'pointer',
    fontWeight: 600,
    boxShadow: 'none',
    transition: 'background-color 0.15s, border-color 0.15s',
    ...style,
  };
}

const CaretButton = ({
  className,
  size = 'sm',
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
      style={caretButtonStyle(size, isOpen, hovered, style)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <CaretLogo variant="teal" size={size === 'md' ? 22 : 21} scale={1.18} />
      Caret AI
    </button>
  );
};

export default CaretButton;
