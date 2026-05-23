import React from 'react';

export const FlowIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Node top */}
    <circle cx="22" cy="10" r="4.5" fill="rgba(255,255,255,0.9)" />
    {/* Node bottom-left */}
    <circle cx="12" cy="32" r="4.5" fill="rgba(255,180,80,1)" />
    {/* Node bottom-right */}
    <circle cx="32" cy="32" r="4.5" fill="rgba(255,180,80,1)" />
    {/* Connector top → bottom-left */}
    <path
      d="M19.5 13.5L14.5 28.5"
      stroke="rgba(255,255,255,0.5)"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeDasharray="2 2"
    />
    {/* Connector top → bottom-right */}
    <path
      d="M24.5 13.5L29.5 28.5"
      stroke="rgba(255,255,255,0.5)"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeDasharray="2 2"
    />
    {/* Connector bottom-left → bottom-right */}
    <path d="M16.5 32H27.5" stroke="rgba(255,180,80,0.7)" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
);
