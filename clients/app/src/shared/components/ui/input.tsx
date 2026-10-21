import React from 'react';
import { cn } from '../../lib/utils';

export type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, style, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          'flex h-10 w-full rounded-lg px-3 py-1 text-base sm:text-sm placeholder:text-[var(--dash-text-muted)] transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-[var(--dash-border-focus)] focus:border-[var(--dash-text)]',
          'disabled:cursor-not-allowed disabled:opacity-50',
          className
        )}
        style={{
          background: 'var(--dash-input-bg)',
          color: 'var(--dash-text)',
          border: '1px solid var(--dash-border)',
          ...style,
        }}
        ref={ref}
        {...props}
      />
    );
  }
);

Input.displayName = 'Input';

export { Input };
