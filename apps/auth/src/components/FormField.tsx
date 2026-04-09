import React, { forwardRef } from 'react';
import { LucideIcon } from 'lucide-react';
import AuthInput from './AuthInput';

export interface FormFieldProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  icon?: LucideIcon;
  error?: string;
  helper?: string;
}

/**
 * Convenience wrapper around AuthInput that adds an optional helper
 * text slot below the field.
 */
const FormField = forwardRef<HTMLInputElement, FormFieldProps>(
  ({ helper, error, ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1.5">
        <AuthInput ref={ref} error={error} {...props} />
        {helper && !error && (
          <p className="text-[11px] text-white/40">{helper}</p>
        )}
      </div>
    );
  }
);

FormField.displayName = 'FormField';

export default FormField;
