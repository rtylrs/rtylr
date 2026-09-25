import React from 'react';

interface SkeletonProps {
  className?: string;
  rows?: number;
}

const Skeleton: React.FC<SkeletonProps> = ({ className = '', rows = 1 }) => {
  if (rows > 1) {
    return (
      <div className={`space-y-2 ${className}`}>
        {Array.from({ length: rows }).map((_, i) => (
          <div
            key={i}
            className="h-4 rounded bg-white/8 animate-pulse"
            style={{ animationDelay: `${i * 100}ms` }}
          />
        ))}
      </div>
    );
  }
  return <div className={`rounded bg-white/8 animate-pulse ${className}`} />;
};

export const SkeletonGrid: React.FC<{ count?: number; className?: string }> = ({
  count = 9,
  className = '',
}) => (
  <div className={`grid grid-cols-3 gap-2.5 ${className}`}>
    {Array.from({ length: count }).map((_, i) => (
      <div
        key={i}
        className="aspect-square rounded-xl bg-white/8 animate-pulse"
        style={{ animationDelay: `${i * 60}ms` }}
      />
    ))}
  </div>
);

export default Skeleton;
