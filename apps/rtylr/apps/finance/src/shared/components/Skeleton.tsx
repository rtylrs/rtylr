import React from 'react';
import { useTranslation } from 'react-i18next';

interface SkeletonProps {
  width?: number | string;
  height?: number | string;
  radius?: number;
  style?: React.CSSProperties;
}

const Skeleton: React.FC<SkeletonProps> = ({ width = '100%', height = 14, radius = 6, style }) => {
  return (
    <div
      style={{
        width,
        height,
        borderRadius: radius,
        background:
          'linear-gradient(90deg, var(--hub-hover) 0%, var(--hub-card) 50%, var(--hub-hover) 100%)',
        backgroundSize: '200% 100%',
        animation: 'finance-skeleton 1.2s linear infinite',
        ...style,
      }}
    />
  );
};

export default Skeleton;

export const SkeletonRow: React.FC<{ count?: number }> = ({ count = 5 }) => {
  const { t } = useTranslation();
  return (
    <div
      role="status"
      aria-live="polite"
      aria-busy="true"
      aria-label={t('common.loadingLabel')}
      style={{ display: 'flex', flexDirection: 'column', gap: 12 }}
    >
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          <Skeleton width={32} height={32} radius={8} />
          <Skeleton width="40%" />
          <Skeleton width="20%" />
          <Skeleton width="15%" />
        </div>
      ))}
    </div>
  );
};
