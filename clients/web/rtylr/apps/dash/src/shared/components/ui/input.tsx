import React from 'react';
import { cn } from '../../lib/utils';

export type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, style, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          'flex h-9 w-full rounded-lg px-3 py-1 text-base sm:text-sm placeholder:text-gray-400 transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent/60',
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
