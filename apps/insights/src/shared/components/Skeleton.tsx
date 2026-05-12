import React from 'react';
import { cn } from '@voxire/ui';

export const SkeletonBlock: React.FC<{ className?: string }> = ({ className = '' }) => (
  <div
    className={cn('animate-pulse rounded-lg', className)}
    style={{ background: 'var(--hub-hover)' }}
  />
);

export const SkeletonRow: React.FC<{ cols?: number }> = ({ cols = 3 }) => (
  <div
    className="flex gap-3 items-center px-3 py-3 rounded-lg"
    style={{ background: 'var(--hub-hover)' }}
  >
    {Array.from({ length: cols }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse flex-1"
        style={{ background: 'var(--hub-active)' }}
      />
    ))}
  </div>
);

export const CardSkeleton: React.FC = () => (
  <div
    className="rounded-xl p-5 space-y-3"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <SkeletonBlock className="h-4 w-24" />
    <SkeletonBlock className="h-8 w-32" />
    <SkeletonBlock className="h-3 w-20" />
  </div>
);
