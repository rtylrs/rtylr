import React from 'react';
import classNames from 'classnames';
import { InputProps } from '@voxire/config';
import './Input.scss';
const EyeIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);

const EyeOffIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
    <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
    <line x1="1" y1="1" x2="23" y2="23" />
  </svg>
);

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
        <div className={classNames('voxire-input__helper', { 'is-error': error })}>
          {helperText}
        </div>
      )}
    </div>
  );
};

export default Input;
