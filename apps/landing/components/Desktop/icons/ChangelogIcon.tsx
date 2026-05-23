import React from 'react';

export const ChangelogIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Clock face */}
    <circle
      cx="22"
      cy="22"
      r="14"
      fill="rgba(255,255,255,0.08)"
      stroke="rgba(255,255,255,0.35)"
      strokeWidth="1"
    />
    <circle cx="22" cy="22" r="1.8" fill="rgba(255,255,255,0.9)" />
    {/* Hour hand */}
    <path d="M22 22L22 12" stroke="rgba(255,255,255,0.9)" strokeWidth="1.5" strokeLinecap="round" />
    {/* Minute hand */}
    <path d="M22 22L29 26" stroke="rgba(255,220,80,1)" strokeWidth="1.5" strokeLinecap="round" />
    {/* Tick marks */}
    {[0, 60, 120, 180, 240, 300].map((deg, i) => {
      const r = (Math.PI * deg) / 180;
      const x1 = 22 + 12 * Math.sin(r);
      const y1 = 22 - 12 * Math.cos(r);
      const x2 = 22 + 10 * Math.sin(r);
      const y2 = 22 - 10 * Math.cos(r);
      return (
        <line
          key={i}
          x1={x1}
          y1={y1}
          x2={x2}
          y2={y2}
          stroke="rgba(255,255,255,0.4)"
          strokeWidth="1"
          strokeLinecap="round"
        />
      );
    })}
  </svg>
);
