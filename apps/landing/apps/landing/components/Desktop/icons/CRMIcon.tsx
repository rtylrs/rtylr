import React from 'react';

export const CRMIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#7c3aed" />
    {/* Connection lines */}
    <line
      x1="22"
      y1="18"
      x2="10"
      y2="30"
      stroke="rgba(180,130,255,0.35)"
      strokeWidth="1"
      strokeDasharray="2 2"
    />
    <line
      x1="22"
      y1="18"
      x2="34"
      y2="30"
      stroke="rgba(180,130,255,0.35)"
      strokeWidth="1"
      strokeDasharray="2 2"
    />
    <line
      x1="10"
      y1="30"
      x2="34"
      y2="30"
      stroke="rgba(180,130,255,0.25)"
      strokeWidth="1"
      strokeDasharray="2 2"
    />
    {/* Center person - head */}
    <circle cx="22" cy="12" r="5.5" fill="rgba(255,255,255,0.88)" />
    {/* Center person - body arc */}
    <path
      d="M12 26c0-5.523 4.477-10 10-10s10 4.477 10 10"
      stroke="rgba(255,255,255,0.88)"
      strokeWidth="2"
      strokeLinecap="round"
      fill="none"
    />
    {/* Left satellite person */}
    <circle cx="9" cy="31" r="3.8" fill="rgba(190,140,255,0.85)" />
    <path
      d="M4 40c0-3.314 2.239-6 5-6"
      stroke="rgba(190,140,255,0.6)"
      strokeWidth="1.5"
      strokeLinecap="round"
      fill="none"
    />
    {/* Right satellite person */}
    <circle cx="35" cy="31" r="3.8" fill="rgba(190,140,255,0.85)" />
    <path
      d="M40 40c0-3.314-2.239-6-5-6"
      stroke="rgba(190,140,255,0.6)"
      strokeWidth="1.5"
      strokeLinecap="round"
      fill="none"
    />
    {/* Node dots at connection points */}
    <circle cx="22" cy="18" r="1.5" fill="rgba(255,255,255,0.5)" />
    <circle cx="10" cy="30" r="1.2" fill="rgba(190,140,255,0.7)" />
    <circle cx="34" cy="30" r="1.2" fill="rgba(190,140,255,0.7)" />
  </svg>
);
