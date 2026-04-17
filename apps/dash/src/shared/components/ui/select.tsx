import React from 'react';
import { cn } from '../../lib/utils';

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {}

const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, children, style, ...props }, ref) => {
    return (
      <select
        ref={ref}
        className={cn(
          'flex h-9 w-full rounded-lg px-3 py-1 text-sm',
          'focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent/60',
          'disabled:cursor-not-allowed disabled:opacity-50 appearance-none',
          className
        )}
        style={{
          background: 'var(--dash-input-bg)',
          color: 'var(--dash-text)',
          border: '1px solid var(--dash-border)',
          ...style,
        }}
        {...props}
      >
        {children}
      </select>
    );
  }
);

Select.displayName = 'Select';

export { Select };
