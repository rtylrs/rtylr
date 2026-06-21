import React from 'react';
import caretTealLogo from '../assets/caret-teal.png';
import caretWhiteLogo from '../assets/caret-white.png';

export interface CaretLogoProps {
  variant?: 'teal' | 'white';
  size?: number;
  scale?: number;
  label?: string;
  style?: React.CSSProperties;
}

export const CaretLogo: React.FC<CaretLogoProps> = ({
  variant = 'teal',
  size = 18,
  scale = 1,
  label,
  style,
}) => {
  const src = variant === 'white' ? caretWhiteLogo : caretTealLogo;

  return (
    <span
      aria-hidden={label ? undefined : true}
      aria-label={label}
      role={label ? 'img' : undefined}
      style={{
        width: size,
        height: size,
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
        flexShrink: 0,
        ...style,
      }}
    >
      <img
        src={src}
        alt=""
        draggable={false}
        style={{
          width: size,
          height: size,
          display: 'block',
          objectFit: 'contain',
          pointerEvents: 'none',
          transform: scale === 1 ? undefined : `scale(${scale})`,
        }}
      />
    </span>
  );
};

export default CaretLogo;
