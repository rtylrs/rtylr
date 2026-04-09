import React from 'react';
import { cn } from '../lib/cn';

interface SpinnerProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

const sizeMap = { sm: 'w-4 h-4', md: 'w-6 h-6', lg: 'w-10 h-10' };

const Spinner: React.FC<SpinnerProps> = ({ className, size = 'md' }) => (
  <span
    className={cn(
      'inline-block rounded-full border-2 border-current border-t-transparent animate-spin-slow',
      sizeMap[size],
      className
    )}
    aria-label="Loading"
  />
);

export { Spinner };
