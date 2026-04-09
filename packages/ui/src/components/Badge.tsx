import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../lib/cn';

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold',
  {
    variants: {
      variant: {
        default:     'bg-primary/10 text-primary',
        secondary:   'bg-secondary/10 text-secondary',
        accent:      'bg-accent/10 text-accent',
        success:     'bg-success/10 text-success',
        destructive: 'bg-destructive/10 text-destructive',
        warning:     'bg-warning/10 text-warning',
        outline:     'border border-current bg-transparent',
      },
    },
    defaultVariants: { variant: 'default' },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

const Badge: React.FC<BadgeProps> = ({ className, variant, ...props }) => (
  <span className={cn(badgeVariants({ variant, className }))} {...props} />
);

export { Badge, badgeVariants };
