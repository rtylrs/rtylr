import React from 'react';

/**
 * Custom vector illustration of a classic restaurant plastic coaster pager:
 * dark puck housing, charging dock pins, red LED perimeter dots and 7-segment display.
 */
export const PlasticPagerSvg = ({ className }: { className?: string }): React.JSX.Element => (
  <svg
    viewBox="0 0 180 140"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    aria-hidden="true"
  >
    {/* Soft drop shadow */}
    <ellipse cx="90" cy="122" rx="58" ry="10" fill="#141B17" fillOpacity="0.1" />
    {/* Charging base plate */}
    <rect x="44" y="96" width="92" height="22" rx="10" fill="#2A312D" />
    <rect x="52" y="102" width="76" height="6" rx="3" fill="#3D4742" />
    {/* Outer plastic coaster puck */}
    <rect
      x="34"
      y="20"
      width="112"
      height="86"
      rx="26"
      fill="#1F2622"
      stroke="#36403B"
      strokeWidth="2"
    />
    <rect x="42" y="28" width="96" height="70" rx="20" fill="#2B332F" />
    {/* Blinking red LED ring */}
    <circle cx="56" cy="40" r="3" fill="#EF4444" fillOpacity="0.9" />
    <circle cx="90" cy="36" r="3" fill="#EF4444" fillOpacity="0.55" />
    <circle cx="124" cy="40" r="3" fill="#EF4444" fillOpacity="0.9" />
    <circle cx="56" cy="86" r="3" fill="#EF4444" fillOpacity="0.55" />
    <circle cx="124" cy="86" r="3" fill="#EF4444" fillOpacity="0.55" />
    {/* LED digit window */}
    <rect
      x="61"
      y="45"
      width="58"
      height="34"
      rx="8"
      fill="#121614"
      stroke="#3E4843"
      strokeWidth="1.5"
    />
    <text
      x="90"
      y="69"
      textAnchor="middle"
      fill="#F87171"
      fontSize="22"
      fontWeight="800"
      fontFamily="ui-monospace, SFMono-Regular, Menlo, monospace"
      letterSpacing="2"
    >
      42
    </text>
    {/* Low battery indicator */}
    <rect x="78" y="85" width="20" height="8" rx="2" stroke="#9CA3AF" strokeWidth="1.4" />
    <rect x="98.5" y="87.5" width="2.5" height="3" rx="0.8" fill="#9CA3AF" />
    <rect x="80.5" y="87" width="5" height="4" rx="1" fill="#EF4444" />
    {/* Subtle retired badge */}
    <circle cx="144" cy="28" r="14" fill="#FEE2E2" stroke="#FCA5A5" strokeWidth="1.5" />
    <path
      d="M139 23L149 33M149 23L139 33"
      stroke="#B91C1C"
      strokeWidth="2.4"
      strokeLinecap="round"
    />
  </svg>
);

/**
 * Custom vector illustration of Rennili's phone-based pager:
 * QR scan badge, forest-green phone frame ringing with sound waves.
 */
export const PhonePagerSvg = ({ className }: { className?: string }): React.JSX.Element => (
  <svg
    viewBox="0 0 180 140"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    aria-hidden="true"
  >
    {/* Soft glow */}
    <ellipse cx="90" cy="124" rx="50" ry="8" fill="#0E4D3B" fillOpacity="0.14" />
    {/* Ringing sound waves */}
    <path
      d="M44 46C36 55 36 77 44 86"
      stroke="#15803D"
      strokeWidth="3"
      strokeLinecap="round"
      strokeOpacity="0.45"
    />
    <path
      d="M32 38C21 51 21 81 32 94"
      stroke="#15803D"
      strokeWidth="3"
      strokeLinecap="round"
      strokeOpacity="0.22"
    />
    <path
      d="M136 46C144 55 144 77 136 86"
      stroke="#15803D"
      strokeWidth="3"
      strokeLinecap="round"
      strokeOpacity="0.45"
    />
    <path
      d="M148 38C159 51 159 81 148 94"
      stroke="#15803D"
      strokeWidth="3"
      strokeLinecap="round"
      strokeOpacity="0.22"
    />
    {/* Smartphone body */}
    <rect x="56" y="14" width="68" height="108" rx="18" fill="#141B17" />
    <rect x="61" y="19" width="58" height="98" rx="14" fill="#15803D" />
    {/* Dynamic island */}
    <rect x="79" y="24" width="22" height="5" rx="2.5" fill="#0A3B2D" />
    {/* Bell icon */}
    <path
      d="M90 38C85.6 38 82 41.6 82 46V51L79.5 54.5C79.1 55.1 79.5 56 80.3 56H99.7C100.5 56 100.9 55.1 100.5 54.5L98 51V46C98 41.6 94.4 38 90 38Z"
      fill="#FFFFFF"
    />
    <circle cx="90" cy="59" r="2.2" fill="#D1FAE5" />
    {/* Order badge */}
    <rect x="70" y="67" width="40" height="24" rx="8" fill="#FFFFFF" fillOpacity="0.18" />
    <text
      x="90"
      y="84"
      textAnchor="middle"
      fill="#FFFFFF"
      fontSize="16"
      fontWeight="800"
      fontFamily="system-ui, -apple-system, sans-serif"
    >
      42
    </text>
    {/* CTA pill */}
    <rect x="70" y="97" width="40" height="11" rx="5.5" fill="#FFFFFF" />
    {/* Floating green check badge */}
    <circle cx="128" cy="24" r="13" fill="#DCFCE7" stroke="#86EFAC" strokeWidth="1.5" />
    <path
      d="M122.5 24.5L126.2 28.2L133.5 20.5"
      stroke="#15803D"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
