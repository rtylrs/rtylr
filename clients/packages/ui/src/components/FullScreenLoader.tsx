import React from 'react';
import { Spinner } from './Spinner';
import { cn } from '../lib/cn';

interface FullScreenLoaderProps {
  className?: string;
}

const FullScreenLoader: React.FC<FullScreenLoaderProps> = ({ className }) => (
  <div
    className={cn(
      'fixed inset-0 z-50 flex items-center justify-center bg-white dark:bg-pos-bg',
      className
    )}
  >
    <Spinner size="lg" className="text-accent" />
  </div>
);

export { FullScreenLoader };
