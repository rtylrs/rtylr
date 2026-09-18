import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../lib/cn';

const badgeVariants = cva('inline-flex items-center rounded-full font-semibold', {
  variants: {
    size: {
      // 10px, so a pill beside 11px body text does not force a 36px table row
      // taller. md is the existing treatment and stays the default.
      sm: 'px-2 py-0 text-[10px]',
      md: 'px-2 py-0.5 text-xs',
    },
    variant: {
      default: 'bg-primary/10 text-primary',
      secondary: 'bg-secondary/10 text-secondary',
      accent: 'bg-accent/10 text-accent',
      success: 'bg-success/10 text-success',
      destructive: 'bg-destructive/10 text-destructive',
      warning: 'bg-warning/10 text-warning',
      outline: 'border border-current bg-transparent',
    },
  },
  defaultVariants: { variant: 'default', size: 'md' },
});

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

const Badge: React.FC<BadgeProps> = ({ className, variant, size, ...props }) => (
  <span className={cn(badgeVariants({ variant, size, className }))} {...props} />
);

export { Badge, badgeVariants };
