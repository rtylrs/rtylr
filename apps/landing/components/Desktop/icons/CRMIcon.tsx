import React from 'react';

export const CRMIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Center person */}
    <circle cx="22" cy="16" r="5" fill="rgba(255,255,255,0.9)" />
    <path
      d="M12 32c0-5.523 4.477-10 10-10s10 4.477 10 10"
      stroke="rgba(255,255,255,0.9)"
      strokeWidth="2"
      strokeLinecap="round"
      fill="none"
    />
    {/* Left person */}
    <circle cx="10" cy="19" r="3.5" fill="rgba(200,160,255,0.7)" />
    <path
      d="M3 33c0-3.866 3.134-7 7-7"
      stroke="rgba(200,160,255,0.6)"
      strokeWidth="1.5"
      strokeLinecap="round"
      fill="none"
    />
    {/* Right person */}
    <circle cx="34" cy="19" r="3.5" fill="rgba(200,160,255,0.7)" />
    <path
      d="M41 33c0-3.866-3.134-7-7-7"
      stroke="rgba(200,160,255,0.6)"
      strokeWidth="1.5"
      strokeLinecap="round"
      fill="none"
    />
  </svg>
);
