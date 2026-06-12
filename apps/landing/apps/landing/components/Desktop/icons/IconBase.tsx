import React from 'react';

interface IconBaseProps {
  size?: number;
  from: string;
  to: string;
  children: React.ReactNode;
}

// True iOS app-icon contour for a 44pt grid: straight edges with
// continuously-smoothed corners.
const SQUIRCLE =
  'M0 15C0 4.4 4.4 0 15 0L29 0C39.6 0 44 4.4 44 15L44 29C44 39.6 39.6 44 29 44L15 44C4.4 44 0 39.6 0 29Z';

// The material stack, bottom to top: base gradient, a radial inner glow
// behind the glyph (the "lit from within" feel), bottom shade, edge light.
// Glyph shapes inherit a white->cool-grey gradient from the group, so
// every glyph reads dimensional instead of flat white. Shapes that need
// detail use explicit rgba knockouts or accent fills.
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
        <radialGradient id={`glow-${uid}`} cx="0.5" cy="0.3" r="0.72">
          <stop offset="0" stopColor="#fff" stopOpacity="0.32" />
          <stop offset="0.55" stopColor="#fff" stopOpacity="0.08" />
          <stop offset="1" stopColor="#fff" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={`shade-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0.55" stopColor="#000" stopOpacity="0" />
          <stop offset="1" stopColor="#000" stopOpacity="0.22" />
        </linearGradient>
        <linearGradient id={`edge-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#fff" stopOpacity="0.55" />
          <stop offset="0.25" stopColor="#fff" stopOpacity="0.16" />
          <stop offset="1" stopColor="#fff" stopOpacity="0.06" />
        </linearGradient>
        <linearGradient id={`glyph-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#ffffff" />
          <stop offset="1" stopColor="#d7dfeb" />
        </linearGradient>
        <filter id={`gs-${uid}`} x="-25%" y="-25%" width="150%" height="150%">
          <feDropShadow dx="0" dy="0.9" stdDeviation="1" floodColor="#000" floodOpacity="0.32" />
        </filter>
      </defs>
      <path d={SQUIRCLE} fill={`url(#bg-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#glow-${uid})`} />
      <path d={SQUIRCLE} fill={`url(#shade-${uid})`} />
      <path d={SQUIRCLE} stroke={`url(#edge-${uid})`} strokeWidth="0.75" fill="none" />
      <g
        filter={`url(#gs-${uid})`}
        fill={`url(#glyph-${uid})`}
        transform="translate(8.5 8.5) scale(1.125)"
      >
        {children}
      </g>
    </svg>
  );
};
