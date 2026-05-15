import React from 'react';
import { caretTokens, caretRadius } from './tokens';

export interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: 'sm' | 'md';
  variant?: 'ghost' | 'soft' | 'solid';
  active?: boolean;
  label: string;
}

const sizes = {
  sm: { box: 24, icon: 13 },
  md: { box: 28, icon: 15 },
};

export const IconButton: React.FC<React.PropsWithChildren<IconButtonProps>> = ({
  size = 'md',
  variant = 'ghost',
  active = false,
  label,
  children,
  style,
  ...rest
}) => {
  const dim = sizes[size];
  const base: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: dim.box,
    height: dim.box,
    borderRadius: caretRadius.sm,
    background:
      variant === 'solid'
        ? caretTokens.accent
        : variant === 'soft' || active
          ? caretTokens.accentSoft
          : 'transparent',
    color:
      variant === 'solid'
        ? '#fff'
        : active
          ? caretTokens.accent
          : caretTokens.textMuted,
    border: 'none',
    cursor: rest.disabled ? 'not-allowed' : 'pointer',
    transition: 'background 0.12s, color 0.12s',
    flexShrink: 0,
  };
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      {...rest}
      style={{ ...base, ...style }}
    >
      {children}
    </button>
  );
};

export default IconButton;
