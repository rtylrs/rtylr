import React from 'react';

export const TerminalIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect
      x="4"
      y="7"
      width="36"
      height="30"
      rx="4"
      fill="rgba(10,10,10,0.9)"
      stroke="rgba(255,255,255,0.25)"
      strokeWidth="1"
    />
    <rect x="4" y="7" width="36" height="8" rx="4" fill="rgba(40,40,44,0.95)" />
    <rect x="4" y="11" width="36" height="4" fill="rgba(40,40,44,0.95)" />
    <circle cx="10" cy="11" r="1.5" fill="#ff5f57" />
    <circle cx="15.5" cy="11" r="1.5" fill="#febc2e" />
    <circle cx="21" cy="11" r="1.5" fill="#28c840" />
    <path
      d="M11 22l5 4-5 4"
      stroke="#22c55e"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <rect x="19" y="29" width="12" height="1.5" rx="0.75" fill="#22c55e" opacity="0.7" />
  </svg>
);
