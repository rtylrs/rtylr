import React from 'react';

export const BlogIcon = ({ size = 44 }: { size?: number }) => {
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
        <linearGradient id={`blog-bg-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#fb7a7a" />
          <stop offset="1" stopColor="#db2626" />
        </linearGradient>
      </defs>
      <rect x="0" y="0" width="44" height="44" rx="10" fill={`url(#blog-bg-${uid})`} />
      {/* RSS wave 1 - outermost */}
      <path
        d="M8 38c0-16.569 13.431-30 30-30"
        stroke="rgba(255,80,60,0.3)"
        strokeWidth="2.2"
        strokeLinecap="round"
        fill="none"
      />
      {/* RSS wave 2 */}
      <path
        d="M8 30c0-12.15 9.85-22 22-22"
        stroke="rgba(255,80,60,0.55)"
        strokeWidth="2.2"
        strokeLinecap="round"
        fill="none"
      />
      {/* RSS wave 3 - innermost */}
      <path
        d="M8 22c0-7.732 6.268-14 14-14"
        stroke="rgba(255,80,60,0.85)"
        strokeWidth="2.2"
        strokeLinecap="round"
        fill="none"
      />
      {/* RSS dot */}
      <circle cx="9.5" cy="38" r="3.5" fill="rgba(255,80,60,1)" />
      {/* Pen / edit icon overlay */}
      <rect
        x="26"
        y="6"
        width="14"
        height="5"
        rx="2"
        transform="rotate(45 26 6)"
        fill="rgba(255,255,255,0.75)"
      />
      <path d="M33 13l3-3 3 3-3 3-3-3Z" fill="rgba(255,255,255,0.5)" />
      <path d="M27 19l-4 4-1 4 4-1 4-4-3-3Z" fill="rgba(255,255,255,0.7)" />
    </svg>
  );
};
