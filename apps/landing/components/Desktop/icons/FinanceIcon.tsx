import React from 'react';

export const FinanceIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Chart bars */}
    <rect x="8" y="28" width="5" height="8" rx="1.5" fill="rgba(255,255,255,0.4)" />
    <rect x="15" y="22" width="5" height="14" rx="1.5" fill="rgba(255,255,255,0.55)" />
    <rect x="22" y="17" width="5" height="19" rx="1.5" fill="rgba(255,255,255,0.7)" />
    <rect x="29" y="12" width="5" height="24" rx="1.5" fill="url(#fin-bar)" />
    {/* Trend line */}
    <path
      d="M10.5 30L17.5 24L24.5 19L31.5 14"
      stroke="rgba(255,220,80,1)"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
    <circle cx="31.5" cy="14" r="2.5" fill="rgba(255,220,80,1)" />
    {/* Dollar sign */}
    <text
      x="8"
      y="12"
      fontSize="9"
      fontWeight="700"
      fill="rgba(255,220,80,1)"
      fontFamily="system-ui"
    >
      $
    </text>
    <defs>
      <linearGradient id="fin-bar" x1="29" y1="12" x2="34" y2="36" gradientUnits="userSpaceOnUse">
        <stop stopColor="#90f050" />
        <stop offset="1" stopColor="#50a020" />
      </linearGradient>
    </defs>
  </svg>
);
