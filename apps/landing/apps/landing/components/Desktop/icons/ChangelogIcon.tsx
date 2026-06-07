import React from 'react';

export const ChangelogIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#0891b2" />
    {/* Outer ring */}
    <circle
      cx="22"
      cy="22"
      r="16"
      fill="rgba(255,255,255,0.06)"
      stroke="rgba(255,255,255,0.28)"
      strokeWidth="1"
    />
    {/* Inner ring */}
    <circle
      cx="22"
      cy="22"
      r="12"
      fill="rgba(255,255,255,0.04)"
      stroke="rgba(255,255,255,0.12)"
      strokeWidth="0.6"
    />
    {/* Tick marks - 12 positions */}
    {Array.from({ length: 12 }).map((_, i) => {
      const angle = (Math.PI * 2 * i) / 12;
      const isMajor = i % 3 === 0;
      const r1 = isMajor ? 13.5 : 14.5;
      const r2 = 16;
      return (
        <line
          key={i}
          x1={22 + r1 * Math.sin(angle)}
          y1={22 - r1 * Math.cos(angle)}
          x2={22 + r2 * Math.sin(angle)}
          y2={22 - r2 * Math.cos(angle)}
          stroke={isMajor ? 'rgba(255,255,255,0.55)' : 'rgba(255,255,255,0.25)'}
          strokeWidth={isMajor ? 1.2 : 0.75}
          strokeLinecap="round"
        />
      );
    })}
    {/* Hour hand - pointing ~10 o'clock */}
    <line
      x1="22"
      y1="22"
      x2="16"
      y2="14"
      stroke="rgba(255,255,255,0.9)"
      strokeWidth="1.75"
      strokeLinecap="round"
    />
    {/* Minute hand - pointing ~2 o'clock */}
    <line
      x1="22"
      y1="22"
      x2="28.5"
      y2="15"
      stroke="rgba(255,210,60,1)"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
    {/* Center cap */}
    <circle cx="22" cy="22" r="2" fill="rgba(255,255,255,0.9)" />
    <circle cx="22" cy="22" r="0.9" fill="#111" />
  </svg>
);
