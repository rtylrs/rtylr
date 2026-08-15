import React from 'react';
import { cn } from '../../lib/utils';

const Skeleton = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('animate-pulse rounded-md bg-black/10 dark:bg-white/10', className)}
      {...props}
    />
  )
);

Skeleton.displayName = 'Skeleton';

export { Skeleton };
