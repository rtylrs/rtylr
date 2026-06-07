import React from 'react';

export const POSIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#2563eb" />
    {/* Monitor body */}
    <rect
      x="6"
      y="6"
      width="32"
      height="22"
      rx="3"
      fill="rgba(255,255,255,0.1)"
      stroke="rgba(255,255,255,0.3)"
      strokeWidth="1"
    />
    {/* Screen face */}
    <rect x="8" y="8" width="28" height="18" rx="2" fill="#0d1117" />
    {/* Screen content - receipt lines */}
    <rect x="11" y="11" width="9" height="1.5" rx="0.75" fill="rgba(255,255,255,0.7)" />
    <rect x="11" y="14" width="6" height="1.5" rx="0.75" fill="rgba(255,255,255,0.4)" />
    <rect x="11" y="17" width="7" height="1.5" rx="0.75" fill="rgba(255,255,255,0.4)" />
    {/* Amount in yellow */}
    <rect x="24" y="11" width="8" height="1.5" rx="0.75" fill="rgba(255,220,80,0.9)" />
    <rect x="24" y="14" width="8" height="1.5" rx="0.75" fill="rgba(255,220,80,0.6)" />
    <rect x="24" y="17" width="8" height="1.5" rx="0.75" fill="rgba(255,220,80,0.6)" />
    {/* Divider line */}
    <line x1="11" y1="20.5" x2="33" y2="20.5" stroke="rgba(255,255,255,0.15)" strokeWidth="0.75" />
    {/* Total row */}
    <rect x="11" y="22.5" width="5" height="1.5" rx="0.75" fill="rgba(255,255,255,0.5)" />
    <rect x="25" y="22.5" width="7" height="1.5" rx="0.75" fill="rgba(80,220,255,0.9)" />
    {/* Stand neck */}
    <rect x="19" y="28" width="6" height="2.5" fill="rgba(255,255,255,0.18)" />
    {/* Stand base */}
    <rect x="13" y="30" width="18" height="2.5" rx="1.25" fill="rgba(255,255,255,0.22)" />
    {/* Card reader */}
    <rect
      x="6"
      y="34"
      width="32"
      height="5"
      rx="2.5"
      fill="rgba(255,255,255,0.08)"
      stroke="rgba(255,255,255,0.2)"
      strokeWidth="0.75"
    />
    {/* Card slot */}
    <rect x="9" y="35.5" width="7" height="2" rx="0.5" fill="rgba(255,220,80,0.5)" />
    {/* Tap icon dots */}
    <circle cx="30" cy="36.5" r="0.8" fill="rgba(80,220,255,0.7)" />
    <circle cx="33" cy="36.5" r="0.8" fill="rgba(80,220,255,0.5)" />
  </svg>
);
