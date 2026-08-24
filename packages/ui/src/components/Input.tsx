import React, { useState } from 'react';
import * as Label from '@radix-ui/react-label';
import { cn } from '../lib/cn';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  wrapperClassName?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, helperText, id, type, wrapperClassName, ...props }, ref) => {
    const [showPassword, setShowPassword] = useState(false);
    const isPassword = type === 'password';
    const inputType = isPassword ? (showPassword ? 'text' : 'password') : type;
    // A caller that passes `label` without `id` produced an unassociated
    // label - htmlFor got undefined and nothing pointed at this input.
    // useId() generates a stable fallback so the association always exists.
    const generatedId = React.useId();
    const inputId = id ?? generatedId;

    return (
      <div className={cn('flex flex-col gap-1.5', wrapperClassName)}>
        {label && (
          <Label.Root
            htmlFor={inputId}
            className="text-xs font-semibold uppercase tracking-wide text-muted"
          >
            {label}
          </Label.Root>
        )}
        <div className="relative">
          <input
            ref={ref}
            id={inputId}
            type={inputType}
            className={cn(
              'w-full rounded-lg border bg-white/[0.06] px-3.5 py-3 text-sm text-inherit outline-none transition-colors placeholder:text-muted/50',
              'border-black/10 focus:border-accent',
              'dark:border-white/10 dark:focus:border-accent',
              isPassword && 'pr-10',
              error && 'border-destructive focus:border-destructive',
              className
            )}
            {...props}
          />
          {isPassword && (
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted hover:text-inherit transition-colors"
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {showPassword ? (
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
              ) : (
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              )}
            </button>
          )}
        </div>
        {error && <p className="text-xs text-destructive">{error}</p>}
        {helperText && !error && <p className="text-xs text-muted">{helperText}</p>}
      </div>
    );
  }
);

Input.displayName = 'Input';

export { Input };
