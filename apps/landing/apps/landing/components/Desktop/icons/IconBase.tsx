import React from 'react';

interface IconBaseProps {
  size?: number;
  from: string;
  to: string;
  stroke?: string;
  children: React.ReactNode;
}

// iOS app-icon squircle (continuous-curvature superellipse) for a 44pt grid.
// Classic Apple approximation: control points at ~25% of the edge.
const SQUIRCLE =
  'M22 0C4.93 0 0 4.93 0 22C0 39.07 4.93 44 22 44C39.07 44 44 39.07 44 22C44 4.93 39.07 0 22 0Z';

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
        <radialGradient id={`sheen-${uid}`} cx="0.5" cy="-0.12" r="1.15">
          <stop offset="0" stopColor="#fff" stopOpacity="0.42" />
          <stop offset="0.55" stopColor="#fff" stopOpacity="0.06" />
          <stop offset="1" stopColor="#fff" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={`shade-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0.55" stopColor="#000" stopOpacity="0" />
          <stop offset="1" stopColor="#000" stopOpacity="0.2" />
        </linearGradient>
        <linearGradient id={`edge-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#fff" stopOpacity="0.5" />
          <stop offset="0.3" stopColor="#fff" stopOpacity="0.14" />
          <stop offset="1" stopColor="#fff" stopOpacity="0.08" />
        </linearGradient>
        <filter id={`glyph-${uid}`} x="-20%" y="-20%" width="140%" height="140%">
          <feDropShadow dx="0" dy="0.6" stdDeviation="0.6" floodColor="#000" floodOpacity="0.35" />
        </filter>
      </defs>
      <path d={SQUIRCLE} fill={`url(#bg-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#sheen-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#shade-${uid})`} />
      <path d={SQUIRCLE} stroke={`url(#edge-${uid})`} strokeWidth="1" fill="none" />
      <g
        filter={`url(#glyph-${uid})`}
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
