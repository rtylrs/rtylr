import React from 'react';

interface IconBaseProps {
  size?: number;
  from: string;
  to: string;
  children: React.ReactNode;
}

// True iOS app-icon contour for a 44pt grid: straight edges with
// continuously-smoothed corners (curves begin at 38.77% of the edge),
// not a full superellipse - this is what makes it read "squarish".
const SQUIRCLE =
  'M0 15C0 4.4 4.4 0 15 0L29 0C39.6 0 44 4.4 44 15L44 29C44 39.6 39.6 44 29 44L15 44C4.4 44 0 39.6 0 29Z';

// Glyphs are authored on a 24pt grid as filled shapes (white fills with
// rgba(0,0,0,x) knockouts for detail) and scaled into the center of the tile.
export const IconBase: React.FC<IconBaseProps> = ({ size = 44, from, to, children }) => {
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
        <linearGradient id={`sheen-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#fff" stopOpacity="0.18" />
          <stop offset="0.4" stopColor="#fff" stopOpacity="0.03" />
          <stop offset="1" stopColor="#fff" stopOpacity="0" />
        </linearGradient>
        <linearGradient id={`shade-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0.6" stopColor="#000" stopOpacity="0" />
          <stop offset="1" stopColor="#000" stopOpacity="0.16" />
        </linearGradient>
        <filter id={`glyph-${uid}`} x="-25%" y="-25%" width="150%" height="150%">
          <feDropShadow dx="0" dy="0.8" stdDeviation="0.9" floodColor="#000" floodOpacity="0.28" />
        </filter>
      </defs>
      <path d={SQUIRCLE} fill={`url(#bg-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#sheen-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#shade-${uid})`} />
      <path
        d={SQUIRCLE}
        stroke="#fff"
        strokeOpacity="0.12"
        strokeWidth="0.75"
        fill="none"
      />
      <g filter={`url(#glyph-${uid})`} transform="translate(8.5 8.5) scale(1.125)">
        {children}
      </g>
    </svg>
  );
};
