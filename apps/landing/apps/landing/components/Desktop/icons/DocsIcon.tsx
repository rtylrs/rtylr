import React from 'react';

export const DocsIcon = ({ size = 44 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 44 44"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="0" y="0" width="44" height="44" rx="10" fill="#2563eb" />
    {/* Shadow page behind */}
    <rect x="14" y="10" width="21" height="27" rx="2.5" fill="rgba(255,255,255,0.07)" />
    {/* Spine */}
    <rect x="9" y="7" width="5.5" height="30" rx="1.5" fill="rgba(255,255,255,0.22)" />
    {/* Main page */}
    <rect x="13" y="7" width="22" height="30" rx="2.5" fill="rgba(255,255,255,0.92)" />
    {/* Page lines */}
    <rect x="17" y="13" width="14" height="2" rx="1" fill="rgba(15,25,50,0.55)" />
    <rect x="17" y="18" width="11" height="1.5" rx="0.75" fill="rgba(15,25,50,0.3)" />
    <rect x="17" y="22" width="14" height="1.5" rx="0.75" fill="rgba(15,25,50,0.3)" />
    <rect x="17" y="26" width="9" height="1.5" rx="0.75" fill="rgba(15,25,50,0.3)" />
    <rect x="17" y="30" width="12" height="1.5" rx="0.75" fill="rgba(15,25,50,0.2)" />
    {/* Bookmark ribbon */}
    <path d="M29 7v11l-3-2.5-3 2.5V7h6Z" fill="rgba(90,160,255,0.9)" />
  </svg>
);
