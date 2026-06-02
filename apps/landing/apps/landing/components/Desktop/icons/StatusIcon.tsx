import React from 'react';

export const StatusIcon = ({ size = 44 }: { size?: number }) => (
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
      fill="rgba(5,20,10,0.95)"
      stroke="rgba(34,197,94,0.3)"
      strokeWidth="1"
    />
    <path
      d="M8 24h4l3-8 4 16 3-12 3 6 3-4 4 2h4"
      stroke="#22c55e"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <circle cx="22" cy="34" r="2" fill="rgba(34,197,94,0.9)" />
    <circle cx="22" cy="34" r="4" stroke="rgba(34,197,94,0.3)" strokeWidth="1" />
  </svg>
);
