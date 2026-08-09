import React from 'react';
import { ASSETS } from '@voxire/config';

interface LogoProps {
  // 'dark' renders a black mark for light backgrounds; 'light' renders a white
  // mark for dark backgrounds.
  variant?: 'dark' | 'light';
  className?: string;
}

// Renders the rtylr mark from ASSETS.logos, tinted via CSS so the single
// (near-white) asset works on either background without a second file.
const Logo: React.FC<LogoProps> = ({ variant = 'dark', className }) => (
  <img
    src={ASSETS.logos.light}
    alt="rtylr"
    className={className}
    draggable={false}
    style={{ filter: variant === 'light' ? 'brightness(0) invert(1)' : 'brightness(0)' }}
  />
);

export default Logo;
