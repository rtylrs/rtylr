import React from 'react';
import ReactSelect, { StylesConfig, components } from 'react-select';
import classNames from 'classnames';
import { SelectProps } from '@voxire/config';
import './Select.scss';

/**
 * A reusable searchable select component that displays options with title and optional description
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
 * @param isSearchable - Whether the dropdown is searchable (default: true)
 */
const Select: React.FC<SelectProps & { isSearchable?: boolean }> = ({
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
  isSearchable = true,
  children,
}) => {
  const selectedOption = options.find((option) => option.value === value);

  const reactSelectOptions = options.map((option) => ({
    value: option.value,
    label: option.title,
    description: option.description,
  }));

  const selectedReactSelectOption = selectedOption
    ? {
        value: selectedOption.value,
        label: selectedOption.title,
        description: selectedOption.description,
      }
    : null;

  const customStyles: StylesConfig = {
    control: (base, state) => ({
      ...base,
      minHeight: '50px',
      backgroundColor: '#e8e4d9',
      borderColor: error ? '#d32f2f' : '#1a1a1a',
      borderWidth: '1px',
      borderRadius: '0',
      boxShadow: state.isFocused ? (error ? '0 0 0 1px #d32f2f' : '0 0 0 1px #1a1a1a') : 'none',
      '&:hover': {
        borderColor: error ? '#d32f2f' : '#1a1a1a',
      },
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.6 : 1,
    }),
    valueContainer: (base) => ({
      ...base,
      padding: '8px 16px',
    }),
    input: (base) => ({
      ...base,
      margin: '0',
      padding: '0',
      color: '#1a1a1a',
    }),
    placeholder: (base) => ({
      ...base,
      color: 'rgba(26, 26, 26, 0.5)',
      fontSize: '1rem',
    }),
    singleValue: (base) => ({
      ...base,
      color: '#1a1a1a',
      fontSize: '1rem',
    }),
    menu: (base) => ({
      ...base,
      backgroundColor: '#f5f3ed',
      borderRadius: '0',
      border: '1px solid #1a1a1a',
      borderTop: 'none',
      marginTop: '-1px',
      boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
      zIndex: 9999,
    }),
    menuList: (base) => ({
      ...base,
      padding: '0',
      maxHeight: '300px',
    }),
    option: (base, state) => ({
      ...base,
      backgroundColor: state.isSelected
        ? 'rgba(26, 26, 26, 0.08)'
        : state.isFocused
          ? 'rgba(26, 26, 26, 0.05)'
          : 'transparent',
      color: '#1a1a1a',
      cursor: 'pointer',
      padding: '14px 16px',
      borderBottom: '1px solid rgba(26, 26, 26, 0.1)',
      fontSize: '1rem',
      '&:last-child': {
        borderBottom: 'none',
      },
      '&:active': {
        backgroundColor: 'rgba(26, 26, 26, 0.08)',
      },
    }),
    indicatorSeparator: () => ({
      display: 'none',
    }),
    dropdownIndicator: (base) => ({
      ...base,
      color: '#1a1a1a',
      padding: '8px 16px',
      '&:hover': {
        color: '#1a1a1a',
      },
    }),
    clearIndicator: (base) => ({
      ...base,
      color: '#1a1a1a',
      padding: '8px',
      '&:hover': {
        color: '#1a1a1a',
      },
    }),
  };

  const CustomOption = (props: any) => {
    return (
      <components.Option {...props}>
        <div>
          <div>{props.data.label}</div>
          {props.data.description && (
            <div
              style={{
                fontSize: '0.875rem',
                color: 'rgba(26, 26, 26, 0.6)',
                marginTop: '4px',
              }}
            >
              {props.data.description}
            </div>
          )}
        </div>
      </components.Option>
    );
  };

  const dropdownClasses = classNames(
    'voxire-select',
    {
      'voxire-select--disabled': disabled,
      'voxire-select--error': error,
    },
    className
  );

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

      <ReactSelect
        inputId={id}
        name={name}
        options={reactSelectOptions}
        value={selectedReactSelectOption}
        onChange={(newValue: any) => {
          if (onChange && newValue) {
            const foundOption = options.find((o) => o.value === newValue.value);
            if (foundOption) {
              onChange(newValue.value, foundOption);
            }
          }
        }}
        onBlur={onBlur}
        onFocus={onFocus}
        placeholder={placeholder}
        isDisabled={disabled}
        isSearchable={isSearchable}
        isClearable={false}
        styles={customStyles}
        components={{ Option: CustomOption }}
        classNamePrefix="voxire-react-select"
      />

      {helperText && (
        <div
          className={classNames('voxire-select__helper', {
            'voxire-select__helper--error': error,
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
