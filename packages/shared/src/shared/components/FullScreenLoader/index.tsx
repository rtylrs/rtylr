import React from 'react';
import LoadingSpinner from '../LoadingSpinner';
import { LoadingSpinnerProps } from '../../types/common';
import './FullScreenLoader.scss';

interface FullScreenLoaderProps extends Omit<LoadingSpinnerProps, 'className'> {
  backgroundColor?: string;
  className?: string;
}

/**
 * Full-screen loading overlay component that centers a LoadingSpinner
 * and covers the entire viewport with a background overlay.
 *
 * @param size - Size of the loading spinner: small, medium, or large
 * @param color - Color of the spinner: primary, secondary, or inherit
 * @param backgroundColor - Background color of the overlay (default: white)
 * @param children - Optional content to display below the spinner
 * @param className - Additional CSS classes
 */
const FullScreenLoader: React.FC<FullScreenLoaderProps> = ({
  size = 'large',
  color = 'primary',
  backgroundColor = 'white',
  children,
  className,
  ...props
}) => {
  return (
    <div
      className={`voxire-fullscreen-loader ${className || ''}`}
      style={{ backgroundColor }}
      {...props}
    >
      <LoadingSpinner size={size} color={color}>
        {children}
      </LoadingSpinner>
    </div>
  );
};

export default FullScreenLoader;
