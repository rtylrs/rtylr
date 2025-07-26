import React from 'react';
import { CircularProgress, Box } from '@mui/material';
import classNames from 'classnames';
import { LoadingSpinnerProps } from '../../types/common';
import './LoadingSpinner.scss';

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'medium',
  color = 'primary',
  className,
  style,
  children,
  ...props
}) => {
  const spinnerClasses = classNames(
    'voxire-loading-spinner',
    `voxire-loading-spinner--${size}`,
    `voxire-loading-spinner--${color}`,
    className
  );

  const getSizeValue = () => {
    switch (size) {
      case 'small':
        return 24;
      case 'large':
        return 48;
      default:
        return 32;
    }
  };

  return (
    <Box className={spinnerClasses} style={style} {...props}>
      <CircularProgress
        size={getSizeValue()}
        color={color}
        className="voxire-loading-spinner__progress"
      />
      {children && (
        <div className="voxire-loading-spinner__content">{children}</div>
      )}
    </Box>
  );
};

export default LoadingSpinner;
