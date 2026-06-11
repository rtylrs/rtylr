import React from 'react';

export const PricingIcon = ({ size = 44 }: { size?: number }) => {
  const uid = React.useId().replace(/[^a-zA-Z0-9_-]/g, '');
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 44 44"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id={`pri-bg-${uid}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#ffa552" />
          <stop offset="1" stopColor="#e8590c" />
        </linearGradient>
      </defs>
      <rect x="0" y="0" width="44" height="44" rx="10" fill={`url(#pri-bg-${uid})`} />
      {/* Tag outline */}
      <path
        d="M9 9h16l11 13-11 13H9V9Z"
        fill="rgba(255,255,255,0.08)"
        stroke="rgba(255,255,255,0.35)"
        strokeWidth="0.9"
        strokeLinejoin="round"
      />
      {/* Tag hole */}
      <circle
        cx="13.5"
        cy="15.5"
        r="2.5"
        fill="rgba(255,255,255,0.0)"
        stroke="rgba(255,255,255,0.45)"
        strokeWidth="1"
      />
      {/* Tiers / pricing rows */}
      <rect x="13" y="21" width="14" height="2" rx="1" fill="rgba(255,255,255,0.75)" />
      <rect x="13" y="25" width="10" height="1.75" rx="0.875" fill="rgba(255,210,60,0.9)" />
      <rect x="13" y="29" width="7" height="1.5" rx="0.75" fill="rgba(255,255,255,0.3)" />
      {/* Star / best-value badge */}
      <circle
        cx="33"
        cy="11"
        r="5"
        fill="rgba(255,210,60,0.18)"
        stroke="rgba(255,210,60,0.6)"
        strokeWidth="0.8"
      />
      <path
        d="M33 8.5l0.75 2h2l-1.6 1.2 0.6 2L33 12.5l-1.75 1.2 0.6-2L30.25 10.5h2L33 8.5Z"
        fill="rgba(255,210,60,1)"
      />
    </svg>
  );
};
