import React from 'react';
import { Button as MuiButton, CircularProgress } from '@mui/material';
import classNames from 'classnames';
import { ButtonProps } from '../../types/common';
import './Button.scss';

const Button: React.FC<ButtonProps> = ({
  variant = 'contained',
  color = 'primary',
  size = 'medium',
  disabled = false,
  loading = false,
  onClick,
  className,
  style,
  children,
  ...props
}) => {
  const buttonClasses = classNames(
    'voxire-button',
    `voxire-button--${variant}`,
    `voxire-button--${color}`,
    `voxire-button--${size}`,
    {
      'voxire-button--loading': loading,
      'voxire-button--disabled': disabled,
    },
    className
  );

  return (
    <MuiButton
      variant={variant}
      color={color}
      size={size}
      disabled={disabled || loading}
      onClick={onClick}
      className={buttonClasses}
      style={style}
      {...props}
    >
      {loading && (
        <CircularProgress
          size={size === 'small' ? 16 : size === 'large' ? 24 : 20}
          className="voxire-button__loader"
        />
      )}
      {children}
    </MuiButton>
  );
};

export default Button;
