import React from 'react';
import { ASSETS } from '@voxire/config';

interface LogoProps {
  // 'dark' renders a black mark for light backgrounds; 'light' renders a white
  // mark for dark backgrounds.
  variant?: 'dark' | 'light';
  className?: string;
}

// Use the supplied light and dark wordmarks directly so their gradients remain intact.
const Logo: React.FC<LogoProps> = ({ variant = 'dark', className }) => (
  <img
    src={variant === 'light' ? ASSETS.logos.light : '/logo-dark.png'}
    alt="rtylr"
    className={className}
    draggable={false}
  />
);

export default Logo;
