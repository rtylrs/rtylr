import React from 'react';
import { caretRadius, caretTokens } from '../ui/tokens';

export interface SkeletonProps {
  width?: number | string;
  height?: number | string;
  radius?: number | string;
}

/** Animated placeholder block while content loads. */
export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = 12,
  radius = caretRadius.sm,
}) => (
  <span
    aria-hidden="true"
    style={{
      display: 'inline-block',
      width,
      height,
      borderRadius: radius,
      background: `linear-gradient(90deg, ${caretTokens.surfaceHigh}, ${caretTokens.borderStrong}, ${caretTokens.surfaceHigh})`,
      backgroundSize: '200% 100%',
      animation: 'caret-shimmer 1.2s linear infinite',
    }}
  />
);

export default Skeleton;
