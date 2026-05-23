import React from 'react';

export const HRIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Badge card */}
    <rect
      x="10"
      y="10"
      width="24"
      height="28"
      rx="3"
      fill="rgba(255,255,255,0.1)"
      stroke="rgba(255,255,255,0.3)"
      strokeWidth="0.75"
    />
    {/* Clip at top */}
    <rect x="17" y="8" width="10" height="5" rx="2" fill="rgba(255,255,255,0.25)" />
    {/* Avatar circle */}
    <circle cx="22" cy="20" r="5" fill="rgba(255,255,255,0.3)" />
    {/* Name lines */}
    <rect x="14" y="28" width="16" height="2" rx="1" fill="rgba(255,255,255,0.75)" />
    <rect x="16" y="32" width="12" height="1.5" rx="0.75" fill="rgba(255,255,255,0.4)" />
    {/* Check mark badge */}
    <circle cx="31" cy="31" r="5" fill="url(#hr-check)" />
    <path
      d="M28.5 31l2 2 3-3"
      stroke="white"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <defs>
      <linearGradient id="hr-check" x1="26" y1="26" x2="36" y2="36" gradientUnits="userSpaceOnUse">
        <stop stopColor="#40e090" />
        <stop offset="1" stopColor="#20a050" />
      </linearGradient>
    </defs>
  </svg>
);
