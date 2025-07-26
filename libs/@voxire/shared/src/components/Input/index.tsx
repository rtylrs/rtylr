import React from 'react';
import { TextField } from '@mui/material';
import classNames from 'classnames';
import { InputProps } from '../../types/common';
import './Input.scss';

const Input: React.FC<InputProps> = ({
  type = 'text',
  placeholder,
  value,
  defaultValue,
  disabled = false,
  required = false,
  error = false,
  helperText,
  onChange,
  onBlur,
  onFocus,
  className,
  style,
  ...props
}) => {
  const inputClasses = classNames(
    'voxire-input',
    {
      'voxire-input--error': error,
      'voxire-input--disabled': disabled,
    },
    className
  );

  return (
    <TextField
      type={type}
      placeholder={placeholder}
      value={value}
      defaultValue={defaultValue}
      disabled={disabled}
      required={required}
      error={error}
      helperText={helperText}
      onChange={onChange}
      onBlur={onBlur}
      onFocus={onFocus}
      className={inputClasses}
      style={style}
      variant="outlined"
      fullWidth
      {...props}
    />
  );
};

export default Input;
