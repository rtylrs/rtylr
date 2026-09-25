import React from 'react';
import { cn } from '@voxire/ui';

interface SkeletonProps {
  className?: string;
  style?: React.CSSProperties;
}

export const Skeleton: React.FC<SkeletonProps> = ({ className = '', style }) => (
  <div
    className={cn('rounded animate-pulse', className)}
    style={{ background: 'var(--hub-border)', opacity: 0.6, ...style }}
  />
);

export const SkeletonCard: React.FC = () => (
  <div
    className="rounded-xl p-4 space-y-3"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <Skeleton className="h-4 w-24" />
    <Skeleton className="h-8 w-16" />
    <Skeleton className="h-3 w-20" />
  </div>
);

export const SkeletonRow: React.FC<{ cols?: number }> = ({ cols = 4 }) => (
  <div
    className="grid px-3 py-2.5 rounded-lg gap-2"
    style={{ gridTemplateColumns: `repeat(${cols}, 1fr)`, background: 'var(--hub-hover)' }}
  >
    {Array.from({ length: cols }).map((_, i) => (
      <Skeleton key={i} className="h-3" />
    ))}
  </div>
);

export default Skeleton;
