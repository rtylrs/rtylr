import React from 'react';

export const PricingIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Tag shape */}
    <path
      d="M10 10h14l10 12-10 12H10V10Z"
      fill="rgba(255,255,255,0.1)"
      stroke="rgba(255,255,255,0.35)"
      strokeWidth="0.75"
    />
    {/* Tag hole */}
    <circle cx="14" cy="16" r="2.5" fill="rgba(255,255,255,0.4)" />
    {/* Price lines */}
    <rect x="14" y="22" width="12" height="2" rx="1" fill="rgba(255,255,255,0.8)" />
    <rect x="14" y="26" width="8" height="1.5" rx="0.75" fill="rgba(255,220,60,1)" />
  </svg>
);
