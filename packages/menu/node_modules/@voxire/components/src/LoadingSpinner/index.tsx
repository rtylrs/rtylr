import React from 'react';
import classNames from 'classnames';
import { LoadingSpinnerProps } from '@voxire/config';
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

  return (
    <div className={spinnerClasses} style={style} {...props}>
      <div className="voxire-spinner-ring"></div>
      {children && <div className="voxire-spinner-text">{children}</div>}
    </div>
  );
};

export default LoadingSpinner;
