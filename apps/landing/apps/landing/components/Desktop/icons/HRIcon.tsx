import React from 'react';

export const HRIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#0a2428" />
    {/* Card body */}
    <rect
      x="8"
      y="9"
      width="28"
      height="30"
      rx="3.5"
      fill="rgba(255,255,255,0.08)"
      stroke="rgba(255,255,255,0.28)"
      strokeWidth="0.9"
    />
    {/* Card stripe at top */}
    <rect x="8" y="9" width="28" height="7" rx="3.5" fill="rgba(255,255,255,0.12)" />
    <rect x="8" y="12.5" width="28" height="3.5" fill="rgba(255,255,255,0.08)" />
    {/* Lanyard clip */}
    <rect
      x="17.5"
      y="6"
      width="9"
      height="6"
      rx="2.5"
      fill="none"
      stroke="rgba(255,255,255,0.3)"
      strokeWidth="1"
    />
    <rect x="20" y="6" width="4" height="4" rx="1" fill="rgba(255,255,255,0.2)" />
    {/* Avatar circle */}
    <circle
      cx="22"
      cy="23"
      r="5.5"
      fill="rgba(255,255,255,0.22)"
      stroke="rgba(255,255,255,0.3)"
      strokeWidth="0.75"
    />
    {/* Avatar silhouette */}
    <circle cx="22" cy="21" r="2.2" fill="rgba(255,255,255,0.7)" />
    <path d="M16.5 27.5c0-3.038 2.462-5.5 5.5-5.5s5.5 2.462 5.5 5.5" fill="rgba(255,255,255,0.5)" />
    {/* Name lines */}
    <rect x="13" y="31" width="18" height="2" rx="1" fill="rgba(255,255,255,0.7)" />
    <rect x="15" y="35" width="14" height="1.5" rx="0.75" fill="rgba(255,255,255,0.35)" />
    {/* Green verified badge */}
    <circle cx="33" cy="33" r="5.5" fill="#16a34a" />
    <path
      d="M30.5 33l2 2 3.5-3.5"
      stroke="white"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
