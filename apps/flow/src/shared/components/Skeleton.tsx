import React from 'react';

const shimmerStyle = `
@keyframes flow-shimmer {
  from { background-position: -200px 0; }
  to   { background-position: calc(200px + 100%) 0; }
}
`;

interface SkeletonProps {
  className?: string;
  width?: string | number;
  height?: string | number;
  rounded?: string;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className = '',
  width,
  height,
  rounded = '4px',
}) => {
  return (
    <>
      <style>{shimmerStyle}</style>
      <div
        className={`flow-shimmer ${className}`}
        style={{
          width,
          height,
          borderRadius: rounded,
          background:
            'linear-gradient(90deg, var(--flow-hover) 0%, var(--flow-active) 50%, var(--flow-hover) 100%)',
          backgroundSize: '200% 100%',
          animation: 'flow-shimmer 1.4s infinite',
        }}
      />
    </>
  );
};

export const TaskRowSkeleton: React.FC = () => (
  <div
    className="flex items-center gap-3 px-4 py-2 border-b"
    style={{ borderColor: 'var(--flow-border)' }}
  >
    <Skeleton width={14} height={14} rounded="50%" />
    <Skeleton width={40} height={10} />
    <Skeleton width={14} height={14} rounded="50%" />
    <Skeleton width="60%" height={12} />
    <Skeleton width={60} height={10} />
    <Skeleton width={22} height={22} rounded="50%" />
  </div>
);

export const CardSkeleton: React.FC = () => (
  <div
    className="p-3.5 rounded-lg"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <Skeleton width={50} height={10} className="mb-2" />
    <Skeleton width="90%" height={14} className="mb-1.5" />
    <Skeleton width="70%" height={14} className="mb-3" />
    <div className="flex items-center gap-2">
      <Skeleton width={18} height={18} rounded="50%" />
      <Skeleton width={40} height={10} />
    </div>
  </div>
);

export default Skeleton;
