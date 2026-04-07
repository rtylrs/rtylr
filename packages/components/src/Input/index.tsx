import React from 'react';
import classNames from 'classnames';
import { InputProps } from '../../types/common';
import './Input.scss';
import EyeIcon from '../../assets/icons/eye.svg?react';
import EyeOffIcon from '../../assets/icons/eye-off.svg?react';

/**
 * @param type - Input type: text, email, password, number, tel, or url (default: text)
 * @param placeholder - Placeholder text for the input
 * @param value - Controlled value of the input
 * @param defaultValue - Default value for uncontrolled input
 * @param disabled - Whether the input is disabled (default: false)
 * @param required - Whether the input is required (default: false)
 * @param error - Whether to show error state styling (default: false)
 * @param helperText - Helper or error text to display below the input
 * @param onChange - Function called when input value changes
 * @param onBlur - Function called when input loses focus
 * @param onFocus - Function called when input gains focus
 * @param id - HTML id attribute for the input
 * @param name - HTML name attribute for the input
 * @param autoComplete - HTML autocomplete attribute
 * @param label - Label text to display above the input
 * @param hasLabel - Whether to render the label (default: true if label is provided)
 * @param labelProps - Additional props for the label element
 * @param autoFocus - Whether to auto-focus the input on mount (default: false)
 * @param className - Additional CSS classes
 * @param style - Additional inline styles
 * @param children - Optional content to render inside the component
 */
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
