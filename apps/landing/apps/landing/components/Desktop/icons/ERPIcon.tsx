import React from 'react';

export const ERPIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#0284c7" />
    {/* Isometric cube - top face */}
    <path d="M22 7L36 14.5V22L22 29.5L8 22V14.5L22 7Z" fill="none" />
    <path
      d="M22 7L36 14.5L22 22L8 14.5L22 7Z"
      fill="rgba(120,190,255,0.22)"
      stroke="rgba(120,190,255,0.55)"
      strokeWidth="0.8"
      strokeLinejoin="round"
    />
    {/* Left face */}
    <path
      d="M8 14.5L22 22V34L8 26.5V14.5Z"
      fill="rgba(255,255,255,0.07)"
      stroke="rgba(255,255,255,0.2)"
      strokeWidth="0.8"
      strokeLinejoin="round"
    />
    {/* Right face */}
    <path
      d="M36 14.5L22 22V34L36 26.5V14.5Z"
      fill="rgba(120,190,255,0.12)"
      stroke="rgba(120,190,255,0.35)"
      strokeWidth="0.8"
      strokeLinejoin="round"
    />
    {/* Center spine */}
    <line x1="22" y1="22" x2="22" y2="34" stroke="rgba(255,255,255,0.25)" strokeWidth="0.75" />
    {/* Glowing top highlight */}
    <path d="M22 7L36 14.5L22 22L8 14.5L22 7Z" fill="url(#erp-glow)" opacity="0.5" />
    {/* Small nodes at corners */}
    <circle cx="22" cy="7" r="1.5" fill="rgba(120,190,255,0.9)" />
    <circle cx="36" cy="14.5" r="1.2" fill="rgba(120,190,255,0.6)" />
    <circle cx="8" cy="14.5" r="1.2" fill="rgba(120,190,255,0.6)" />
    <circle cx="22" cy="34" r="1.2" fill="rgba(120,190,255,0.5)" />
    <defs>
      <linearGradient id="erp-glow" x1="8" y1="7" x2="36" y2="22" gradientUnits="userSpaceOnUse">
        <stop stopColor="#80c8ff" stopOpacity="0.7" />
        <stop offset="1" stopColor="#4090d0" stopOpacity="0" />
      </linearGradient>
    </defs>
  </svg>
);
