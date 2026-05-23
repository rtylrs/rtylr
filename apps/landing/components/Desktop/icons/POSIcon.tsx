import React from 'react';

export const POSIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Screen */}
    <rect
      x="8"
      y="7"
      width="28"
      height="19"
      rx="2.5"
      fill="rgba(255,255,255,0.15)"
      stroke="rgba(255,255,255,0.35)"
      strokeWidth="1"
    />
    <rect x="10" y="9" width="24" height="2" rx="1" fill="rgba(255,255,255,0.2)" />
    {/* Receipt lines */}
    <rect x="13" y="13" width="10" height="1.5" rx="0.75" fill="rgba(255,255,255,0.75)" />
    <rect x="13" y="16.5" width="7" height="1.5" rx="0.75" fill="rgba(255,255,255,0.5)" />
    <rect x="26" y="13" width="6" height="1.5" rx="0.75" fill="rgba(255,240,100,1)" />
    <rect x="26" y="16.5" width="6" height="1.5" rx="0.75" fill="rgba(255,240,100,1)" />
    {/* Stand */}
    <rect x="18" y="26" width="8" height="3" fill="rgba(255,255,255,0.2)" />
    <rect x="14" y="29" width="16" height="2.5" rx="1.25" fill="rgba(255,255,255,0.25)" />
    {/* Card reader */}
    <rect
      x="8"
      y="33"
      width="28"
      height="5"
      rx="2"
      fill="rgba(255,255,255,0.12)"
      stroke="rgba(255,255,255,0.25)"
      strokeWidth="0.75"
    />
    <rect x="11" y="34.5" width="6" height="2" rx="0.5" fill="rgba(255,240,100,0.7)" />
  </svg>
);
