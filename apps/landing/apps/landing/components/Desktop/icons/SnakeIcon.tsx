import React from 'react';

export const SnakeIcon = ({ size = 44 }: { size?: number }) => {
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
        <linearGradient id={`snake-bg-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#13301a" />
          <stop offset="1" stopColor="#07160a" />
        </linearGradient>
      </defs>
      <rect x="0" y="0" width="44" height="44" rx="10" fill={`url(#snake-bg-${uid})`} />
      <rect x="9" y="9" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.9)" />
      <rect x="15" y="9" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="21" y="9" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="21" y="15" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="21" y="21" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="15" y="21" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="9" y="21" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.7)" />
      <rect x="9" y="27" width="6" height="6" rx="1.5" fill="rgba(34,197,94,0.5)" />
      <circle cx="33" cy="12" r="3" fill="rgba(239,68,68,0.9)" />
    </svg>
  );
};
