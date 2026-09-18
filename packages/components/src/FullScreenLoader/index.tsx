import React from 'react';
import LoadingSpinner from '../LoadingSpinner';
import { LoadingSpinnerProps } from '@voxire/config';
import './FullScreenLoader.scss';

interface FullScreenLoaderProps extends Omit<LoadingSpinnerProps, 'className'> {
  backgroundColor?: string;
  className?: string;
}

const FullScreenLoader: React.FC<FullScreenLoaderProps> = ({
  size = 'large',
  color = 'primary',
  backgroundColor = '#ffffff',
  children,
  className,
  ...props
}) => {
  return (
    <div className={`voxire-fullscreen-loader ${className || ''}`} style={{ backgroundColor }}>
      <LoadingSpinner size={size} color={color}>
        {children}
      </LoadingSpinner>
    </div>
  );
};

export default FullScreenLoader;
