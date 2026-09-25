import React from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '../../lib/utils';

export type SelectProps = React.SelectHTMLAttributes<HTMLSelectElement>;

const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, children, style, ...props }, ref) => {
    return (
      <div className="relative w-full">
        <select
          ref={ref}
          className={cn(
            'flex h-9 w-full rounded-lg ps-3 pe-8 py-1 text-sm',
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
        <ChevronDown
          className="absolute end-2.5 top-1/2 -translate-y-1/2 h-4 w-4 pointer-events-none"
          style={{ color: 'var(--dash-text-dim)' }}
        />
      </div>
    );
  }
);

Select.displayName = 'Select';

export { Select };
