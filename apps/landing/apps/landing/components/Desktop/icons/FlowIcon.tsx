import React from 'react';

export const FlowIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="flow-bg" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0" stopColor="#8a92fb" />
        <stop offset="1" stopColor="#4f46e5" />
      </linearGradient>
    </defs>
    <rect x="0" y="0" width="44" height="44" rx="10" fill="url(#flow-bg)" />
    {/* Connector lines */}
    <path
      d="M22 12.5V20"
      stroke="rgba(255,255,255,0.35)"
      strokeWidth="1.25"
      strokeLinecap="round"
    />
    <path
      d="M18 24L13 30.5"
      stroke="rgba(255,160,60,0.6)"
      strokeWidth="1.25"
      strokeLinecap="round"
    />
    <path
      d="M26 24L31 30.5"
      stroke="rgba(255,160,60,0.6)"
      strokeWidth="1.25"
      strokeLinecap="round"
    />
    <line
      x1="16"
      y1="34"
      x2="28"
      y2="34"
      stroke="rgba(255,160,60,0.4)"
      strokeWidth="1.25"
      strokeLinecap="round"
      strokeDasharray="2 2"
    />
    {/* Top trigger node - circle */}
    <circle cx="22" cy="9" r="4.5" fill="rgba(255,255,255,0.9)" />
    <circle cx="22" cy="9" r="2" fill="rgba(30,30,40,1)" />
    {/* Middle decision diamond */}
    <rect
      x="18"
      y="20"
      width="8"
      height="8"
      rx="1.5"
      transform="rotate(45 22 24)"
      fill="rgba(255,255,255,0.18)"
      stroke="rgba(255,255,255,0.45)"
      strokeWidth="0.9"
    />
    {/* Bottom-left action node */}
    <circle cx="12" cy="34" r="4.5" fill="rgba(255,160,60,1)" />
    <rect x="10" y="32.5" width="4" height="3" rx="0.75" fill="rgba(30,15,0,0.8)" />
    {/* Bottom-right action node */}
    <circle cx="32" cy="34" r="4.5" fill="rgba(255,160,60,1)" />
    <path
      d="M30 34l1.5-1.5 2 2-2 2-1.5-1.5"
      stroke="rgba(30,15,0,0.8)"
      strokeWidth="1"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);
