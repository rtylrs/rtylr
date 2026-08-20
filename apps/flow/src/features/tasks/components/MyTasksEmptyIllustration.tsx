import React from 'react';

const MyTasksEmptyIllustration: React.FC = () => (
  <svg
    width="96"
    height="96"
    viewBox="0 0 96 96"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <rect x="16" y="10" width="64" height="78" rx="8" fill="var(--flow-hover)" />
    <rect
      x="30"
      y="4"
      width="36"
      height="14"
      rx="5"
      fill="var(--flow-card)"
      stroke="var(--flow-border)"
      strokeWidth="1.5"
    />
    <rect x="28" y="30" width="40" height="5" rx="2.5" fill="var(--flow-border)" />
    <rect x="28" y="42" width="30" height="5" rx="2.5" fill="var(--flow-border)" />
    <circle cx="48" cy="66" r="18" fill="var(--accent)" fillOpacity="0.15" />
    <path
      d="M40 66l5.5 5.5L56 60.5"
      stroke="var(--accent)"
      strokeWidth="3.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

export default MyTasksEmptyIllustration;
