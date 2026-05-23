import React from 'react';

export const ERPIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Box body */}
    <path
      d="M22 9L35 15.5V28.5L22 35L9 28.5V15.5L22 9Z"
      fill="rgba(255,255,255,0.08)"
      stroke="rgba(255,255,255,0.3)"
      strokeWidth="0.75"
    />
    {/* Box top face */}
    <path d="M22 9L35 15.5L22 22L9 15.5L22 9Z" fill="rgba(160,210,255,0.25)" />
    {/* Box left face */}
    <path d="M9 15.5L22 22V35L9 28.5V15.5Z" fill="rgba(255,255,255,0.07)" />
    {/* Box right face */}
    <path d="M35 15.5L22 22V35L35 28.5V15.5Z" fill="rgba(255,255,255,0.14)" />
    {/* Center line */}
    <line x1="22" y1="22" x2="22" y2="35" stroke="rgba(255,255,255,0.3)" strokeWidth="0.75" />
    {/* Top highlight */}
    <path d="M22 9L35 15.5L22 22L9 15.5L22 9Z" fill="url(#erp-top)" opacity="0.7" />
    <defs>
      <linearGradient id="erp-top" x1="9" y1="9" x2="35" y2="22" gradientUnits="userSpaceOnUse">
        <stop stopColor="#80c0ff" stopOpacity="0.6" />
        <stop offset="1" stopColor="#4080c0" stopOpacity="0.1" />
      </linearGradient>
    </defs>
  </svg>
);
