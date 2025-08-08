import React from 'react';
import classNames from 'classnames';
import { InputProps } from '../../types/common';
import './Input.scss';
import EyeIcon from '../../assets/icons/eye.svg?react';
import EyeOffIcon from '../../assets/icons/eye-off.svg?react';

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
  id,
  name,
  autoComplete,
  autoFocus = false,
  label,
  hasLabel = true,
  labelProps,
  ...props
}) => {
  const isPassword = type === 'password';
  const [showPassword, setShowPassword] = React.useState(false);

  const containerClasses = classNames('voxire-input', {
    'voxire-input--error': error,
    'voxire-input--disabled': disabled,
  });

  const inputClasses = classNames('voxire-input__field', className);

  return (
    <div style={style} className={containerClasses}>
      {hasLabel && label && (
        <label htmlFor={id} className="voxire-input__label" {...labelProps}>
          {label}
        </label>
      )}
      <input
        id={id}
        name={name}
        className={inputClasses}
        type={isPassword ? (showPassword ? 'text' : 'password') : type}
        placeholder={placeholder}
        value={value}
        defaultValue={defaultValue}
        disabled={disabled}
        required={required}
        onChange={onChange}
        onBlur={onBlur}
        onFocus={onFocus}
        autoComplete={autoComplete}
        autoFocus={autoFocus}
        {...props}
      />
      {isPassword && (
        <button
          type="button"
          className="voxire-input__toggle"
          aria-label={showPassword ? 'Hide password' : 'Show password'}
          onClick={() => setShowPassword((s) => !s)}
        >
          {showPassword ? <EyeOffIcon /> : <EyeIcon />}
        </button>
      )}
      {helperText && (
        <div
          className={classNames('voxire-input__helper', { 'is-error': error })}
        >
          {helperText}
        </div>
      )}
    </div>
  );
};

export default Input;
