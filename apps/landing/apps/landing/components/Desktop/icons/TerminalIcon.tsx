import React from 'react';

export const TerminalIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="term-bg" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0" stopColor="#2a2a32" />
        <stop offset="1" stopColor="#141419" />
      </linearGradient>
    </defs>
    <rect x="0" y="0" width="44" height="44" rx="10" fill="url(#term-bg)" />
    <rect x="0" y="0" width="44" height="11" rx="10" fill="rgba(255,255,255,0.07)" />
    <rect x="0" y="7" width="44" height="4" fill="rgba(255,255,255,0.07)" />
    <circle cx="10" cy="9" r="1.8" fill="#ff5f57" />
    <circle cx="16" cy="9" r="1.8" fill="#febc2e" />
    <circle cx="22" cy="9" r="1.8" fill="#28c840" />
    <path
      d="M12 23l6 5-6 5"
      stroke="#22c55e"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <rect x="22" y="31" width="12" height="2" rx="1" fill="#22c55e" opacity="0.7" />
  </svg>
);
