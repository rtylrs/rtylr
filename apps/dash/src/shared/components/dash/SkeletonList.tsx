import React from 'react';
import { Skeleton } from '../ui';

export const SkeletonList: React.FC<{ rows?: number; withAvatar?: boolean }> = ({
  rows = 5,
  withAvatar = true,
}) => {
  return (
    <div className="space-y-1">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 px-4 py-3 rounded-lg">
          {withAvatar && <Skeleton className="h-9 w-9 rounded-full shrink-0" />}
          <div className="flex-1 min-w-0 space-y-2">
            <Skeleton className="h-3.5 w-1/3" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-6 w-16 rounded-full shrink-0" />
        </div>
      ))}
    </div>
  );
};

export const SkeletonGrid: React.FC<{ cols?: number; rows?: number }> = ({
  cols = 3,
  rows = 2,
}) => {
  return (
    <div
      className={`grid gap-4`}
      style={{ gridTemplateColumns: `repeat(${cols}, minmax(0, 1fr))` }}
    >
      {Array.from({ length: cols * rows }).map((_, i) => (
        <div
          key={i}
          className="rounded-xl p-5 flex flex-col gap-4"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <Skeleton className="h-11 w-11 rounded-xl" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-1/2" />
          </div>
        </div>
      ))}
    </div>
  );
};

export default SkeletonList;
