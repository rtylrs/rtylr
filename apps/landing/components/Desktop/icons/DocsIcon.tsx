import React from 'react';

export const DocsIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Book spine */}
    <rect x="10" y="8" width="5" height="28" rx="1.5" fill="rgba(255,255,255,0.25)" />
    {/* Pages */}
    <rect x="14" y="8" width="20" height="28" rx="2" fill="rgba(255,255,255,0.88)" />
    {/* Text lines */}
    <rect x="18" y="14" width="12" height="2" rx="1" fill="rgba(10,20,40,0.55)" />
    <rect x="18" y="19" width="10" height="1.5" rx="0.75" fill="rgba(10,20,40,0.35)" />
    <rect x="18" y="23" width="12" height="1.5" rx="0.75" fill="rgba(10,20,40,0.35)" />
    <rect x="18" y="27" width="8" height="1.5" rx="0.75" fill="rgba(10,20,40,0.35)" />
    {/* Bookmark */}
    <path d="M28 8v8l-2.5-2-2.5 2V8h5Z" fill="rgba(100,160,255,0.9)" />
  </svg>
);
