import React from 'react';
import { Check } from 'lucide-react';
import { cn } from '../../lib/utils';

export interface PermissionCheckboxProps {
  checked: boolean;
  disabled?: boolean;
  onChange?: (value: boolean) => void;
  label?: string;
  ariaLabel?: string;
  className?: string;
}

export const PermissionCheckbox: React.FC<PermissionCheckboxProps> = ({
  checked,
  disabled,
  onChange,
  label,
  ariaLabel,
  className,
}) => {
  return (
    <label
      className={cn(
        'inline-flex items-center gap-2 cursor-pointer select-none',
        disabled && 'cursor-not-allowed opacity-50',
        className
      )}
    >
      <button
        type="button"
        disabled={disabled}
        onClick={() => onChange?.(!checked)}
        className="h-4 w-4 rounded flex items-center justify-center transition-all shrink-0"
        style={{
          background: checked ? 'var(--accent)' : 'var(--dash-hover)',
          border: `1px solid ${checked ? 'var(--accent)' : 'var(--dash-border)'}`,
        }}
        aria-checked={checked}
        aria-label={ariaLabel || label || 'Permission toggle'}
        role="checkbox"
      >
        {checked && <Check size={11} color="white" strokeWidth={3} />}
      </button>
      {label && (
        <span className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
          {label}
        </span>
      )}
    </label>
  );
};

export default PermissionCheckbox;
