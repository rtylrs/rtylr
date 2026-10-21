import React from 'react';
import { cn } from '../../lib/utils';

type BadgeVariant = 'default' | 'secondary' | 'success' | 'warning' | 'destructive' | 'outline';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-primary/10 text-primary',
  secondary: 'bg-black/8 text-gray-500',
  success: 'bg-emerald-500/15 text-emerald-600 border border-emerald-500/20',
  warning: 'bg-amber-500/15 text-amber-600 border border-amber-500/20',
  destructive: 'bg-red-500/15 text-red-600 border border-red-500/20',
  outline: 'border border-[var(--dash-border)] text-[var(--dash-text-muted)] bg-transparent',
};

const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ variant = 'default', className, ...props }, ref) => (
    <span
      ref={ref}
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        variantClasses[variant],
        className
      )}
      {...props}
    />
  )
);

Badge.displayName = 'Badge';

export { Badge };
export type { BadgeProps, BadgeVariant };
