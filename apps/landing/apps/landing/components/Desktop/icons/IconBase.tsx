import React from 'react';

interface IconBaseProps {
  size?: number;
  from: string;
  to: string;
  stroke?: string;
  children: React.ReactNode;
}

export const IconBase: React.FC<IconBaseProps> = ({
  size = 44,
  from,
  to,
  stroke = '#fff',
  children,
}) => {
  const uid = React.useId().replace(/[^a-zA-Z0-9_-]/g, '');
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 44 44"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id={`bg-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor={from} />
          <stop offset="1" stopColor={to} />
        </linearGradient>
        <linearGradient id={`hl-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#fff" stopOpacity="0.25" />
          <stop offset="0.45" stopColor="#fff" stopOpacity="0.05" />
          <stop offset="1" stopColor="#fff" stopOpacity="0" />
        </linearGradient>
      </defs>
      <rect width="44" height="44" rx="11" fill={`url(#bg-${uid})`} />
      <rect width="44" height="44" rx="11" fill={`url(#hl-${uid})`} />
      <rect x="0.5" y="0.5" width="43" height="43" rx="10.5" stroke="#fff" strokeOpacity="0.16" />
      <g
        transform="translate(8 8) scale(1.1667)"
        stroke={stroke}
        strokeWidth="1.9"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      >
        {children}
      </g>
    </svg>
  );
};
