import React from 'react';

export const DoomIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect
      x="4"
      y="4"
      width="36"
      height="36"
      rx="8"
      fill="rgba(20,5,5,0.95)"
      stroke="rgba(239,68,68,0.3)"
      strokeWidth="1"
    />
    <circle
      cx="22"
      cy="18"
      r="9"
      fill="rgba(239,68,68,0.15)"
      stroke="rgba(239,68,68,0.5)"
      strokeWidth="1"
    />
    <circle cx="18.5" cy="16" r="2" fill="rgba(239,68,68,0.9)" />
    <circle cx="25.5" cy="16" r="2" fill="rgba(239,68,68,0.9)" />
    <path
      d="M17 21c1 2 7 2 8 0"
      stroke="rgba(239,68,68,0.8)"
      strokeWidth="1.2"
      strokeLinecap="round"
    />
    <path
      d="M14 10l4 4M30 10l-4 4"
      stroke="rgba(239,68,68,0.6)"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
    <path d="M10 30h24" stroke="rgba(239,68,68,0.3)" strokeWidth="1" />
    <rect
      x="12"
      y="32"
      width="6"
      height="5"
      rx="1"
      fill="rgba(239,68,68,0.2)"
      stroke="rgba(239,68,68,0.4)"
      strokeWidth="0.75"
    />
    <rect
      x="22"
      y="32"
      width="10"
      height="5"
      rx="1"
      fill="rgba(239,68,68,0.2)"
      stroke="rgba(239,68,68,0.4)"
      strokeWidth="0.75"
    />
  </svg>
);
