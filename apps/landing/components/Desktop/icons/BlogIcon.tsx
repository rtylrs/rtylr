import React from 'react';

export const BlogIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* RSS circles */}
    <circle cx="11" cy="33" r="3" fill="rgba(255,100,80,1)" />
    <path
      d="M11 25c7.18 0 13 5.82 13 13"
      stroke="rgba(255,100,80,0.8)"
      strokeWidth="2"
      strokeLinecap="round"
      fill="none"
    />
    <path
      d="M11 17c11.598 0 21 9.402 21 21"
      stroke="rgba(255,100,80,0.45)"
      strokeWidth="2"
      strokeLinecap="round"
      fill="none"
    />
    {/* Edit pen */}
    <path d="M30 8l6 6-14 14-7 1 1-7 14-14Z" fill="rgba(255,255,255,0.82)" />
    <path d="M30 8l6 6" stroke="rgba(255,255,255,0.4)" strokeWidth="1" />
  </svg>
);
