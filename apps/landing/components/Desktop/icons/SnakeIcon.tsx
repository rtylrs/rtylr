import React from 'react';

export const SnakeIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect
      x="4"
      y="4"
      width="36"
      height="36"
      rx="8"
      fill="rgba(5,15,5,0.95)"
      stroke="rgba(34,197,94,0.3)"
      strokeWidth="1"
    />
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
