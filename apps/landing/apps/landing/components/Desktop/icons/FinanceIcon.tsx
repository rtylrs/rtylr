import React from 'react';

export const FinanceIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="fin-bg" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0" stopColor="#3ddc8a" />
        <stop offset="1" stopColor="#16a34a" />
      </linearGradient>
    </defs>
    <rect x="0" y="0" width="44" height="44" rx="10" fill="url(#fin-bg)" />
    {/* Grid baseline */}
    <line x1="7" y1="36" x2="38" y2="36" stroke="rgba(255,255,255,0.15)" strokeWidth="0.75" />
    <line x1="7" y1="29" x2="38" y2="29" stroke="rgba(255,255,255,0.07)" strokeWidth="0.5" />
    <line x1="7" y1="22" x2="38" y2="22" stroke="rgba(255,255,255,0.07)" strokeWidth="0.5" />
    {/* Bars */}
    <rect x="8" y="30" width="5" height="6" rx="1.5" fill="rgba(255,255,255,0.35)" />
    <rect x="15" y="25" width="5" height="11" rx="1.5" fill="rgba(255,255,255,0.5)" />
    <rect x="22" y="19" width="5" height="17" rx="1.5" fill="rgba(255,255,255,0.65)" />
    <rect x="29" y="13" width="5" height="23" rx="1.5" fill="url(#fin-bar)" />
    {/* Trend line */}
    <polyline
      points="10.5,31 17.5,26 24.5,20 31.5,14"
      stroke="rgba(255,210,60,1)"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
    {/* Trend dot */}
    <circle cx="31.5" cy="14" r="2.5" fill="rgba(255,210,60,1)" />
    <circle cx="31.5" cy="14" r="1" fill="white" />
    {/* Dollar badge */}
    <circle
      cx="10"
      cy="10"
      r="5"
      fill="rgba(255,210,60,0.15)"
      stroke="rgba(255,210,60,0.5)"
      strokeWidth="0.75"
    />
    <text
      x="10"
      y="13.5"
      fontSize="7"
      fontWeight="800"
      fill="rgba(255,210,60,1)"
      fontFamily="system-ui"
      textAnchor="middle"
    >
      $
    </text>
    <defs>
      <linearGradient id="fin-bar" x1="29" y1="13" x2="34" y2="36" gradientUnits="userSpaceOnUse">
        <stop stopColor="#6ee74a" />
        <stop offset="1" stopColor="#3a9920" />
      </linearGradient>
    </defs>
  </svg>
);
