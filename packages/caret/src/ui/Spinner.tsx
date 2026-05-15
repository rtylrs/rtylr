import React from 'react';

export interface SpinnerProps {
  size?: number;
  color?: string;
}

export const Spinner: React.FC<SpinnerProps> = ({ size = 14, color = 'currentColor' }) => {
  return (
    <span
      aria-hidden="true"
      style={{
        display: 'inline-block',
        width: size,
        height: size,
        border: `2px solid ${color}`,
        borderTopColor: 'transparent',
        borderRadius: '50%',
        animation: 'caret-spin 0.7s linear infinite',
      }}
    />
  );
};

export default Spinner;
