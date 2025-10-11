import React, { useState, useRef, useEffect } from 'react';
import classNames from 'classnames';
import { SelectProps } from '../../types/common';
import './Select.scss';

/**
 * A reusable select component that displays options with title and optional description
 * @param options - Array of select options with id, title, optional description, and value
 * @param value - Currently selected value
 * @param placeholder - Placeholder text when no option is selected
 * @param disabled - Whether the dropdown is disabled
 * @param error - Whether to show error state styling
 * @param helperText - Helper or error text to display below the dropdown
 * @param onChange - Function called when selection changes
 * @param onBlur - Function called when dropdown loses focus
 * @param onFocus - Function called when dropdown gains focus
 * @param id - HTML id attribute for the dropdown
 * @param name - HTML name attribute for the dropdown
 * @param label - Label text to display above the dropdown
 * @param hasLabel - Whether to render the label
 * @param labelProps - Additional props for the label element
 * @param required - Whether the field is required
 * @param className - Additional CSS classes
 * @param style - Additional inline styles
 */
const Select: React.FC<SelectProps> = ({
  options = [],
  value,
  placeholder = 'Select an option',
  disabled = false,
  error = false,
  helperText,
  onChange,
  onBlur,
  onFocus,
  id,
  name,
  label,
  hasLabel = label ? true : false,
  labelProps,
  required = false,
  className,
  style,
  children,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);

  const selectedOption = options.find(option => option.value === value);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Close dropdown on escape key
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen) {
        setIsOpen(false);
        toggleRef.current?.focus();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      return () => document.removeEventListener('keydown', handleKeyDown);
    }
  }, [isOpen]);

  const handleToggle = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
      if (!isOpen && onFocus) {
        onFocus();
      }
    }
  };

  const handleSelect = (option: typeof options[0]) => {
    if (!disabled) {
      setIsOpen(false);
      if (onChange) {
        onChange(option.value, option);
      }
      if (onBlur) {
        onBlur();
      }
    }
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLButtonElement>) => {
    if (disabled) return;
    
    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        handleToggle();
        break;
      case 'ArrowDown':
        event.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
        }
        break;
    }
  };

  const dropdownClasses = classNames(
    'voxire-select',
    {
      'voxire-select--disabled': disabled,
      'voxire-select--error': error,
      'voxire-select--open': isOpen,
    },
    className
  );

  const toggleClasses = classNames('voxire-select__toggle', {
    'voxire-select__toggle--placeholder': !selectedOption,
    'voxire-select__toggle--disabled': disabled,
    'voxire-select__toggle--error': error,
  });

  return (
    <div className={dropdownClasses} style={style}>
      {hasLabel && label && (
        <label 
          {...labelProps}
          htmlFor={id}
          className={classNames('voxire-select__label', labelProps?.className)}
        >
          {label}
          {required && <span className="voxire-select__required">*</span>}
        </label>
      )}
      
      <div className="voxire-select__container" ref={dropdownRef}>
        <button
          ref={toggleRef}
          type="button"
          id={id}
          name={name}
          className={toggleClasses}
          onClick={handleToggle}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          aria-haspopup="listbox"
          aria-expanded={isOpen}
          aria-labelledby={hasLabel && label ? `${id}-label` : undefined}
        >
          <span className="voxire-select__value">
            {selectedOption ? selectedOption.title : placeholder}
          </span>
          <svg 
            className={classNames('voxire-select__chevron', {
              'voxire-select__chevron--open': isOpen
            })}
            width="12" 
            height="12" 
            viewBox="0 0 12 12" 
            fill="none" 
            xmlns="http://www.w3.org/2000/svg"
          >
            <path 
              d="M3 4.5L6 7.5L9 4.5" 
              stroke="currentColor" 
              strokeWidth="1.5" 
              strokeLinecap="round" 
              strokeLinejoin="round"
            />
          </svg>
        </button>
        
        {isOpen && (
          <div 
            className="voxire-select__menu"
            role="listbox"
            aria-labelledby={hasLabel && label ? `${id}-label` : undefined}
          >
            {options.map((option) => (
              <div
                key={option.id}
                className={classNames('voxire-select__option', {
                  'voxire-select__option--selected': option.value === value
                })}
                onClick={() => handleSelect(option)}
                role="option"
                aria-selected={option.value === value}
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleSelect(option);
                  }
                }}
              >
                <div className="voxire-select__option-title">
                  {option.title}
                </div>
                {option.description && (
                  <div className="voxire-select__option-description">
                    {option.description}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
      
      {helperText && (
        <div 
          className={classNames('voxire-select__helper', {
            'voxire-select__helper--error': error
          })}
        >
          {helperText}
        </div>
      )}
      
      {children}
    </div>
  );
};

export default Select;