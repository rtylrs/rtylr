import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../lib/cn';

const alertVariants = cva('rounded-lg border px-4 py-3 text-sm', {
  variants: {
    variant: {
      default: 'border-border bg-muted/10 text-inherit',
      destructive: 'border-destructive/30 bg-destructive/10 text-destructive',
      success: 'border-success/30 bg-success/10 text-success',
      warning: 'border-warning/30 bg-warning/10 text-warning',
    },
  },
  defaultVariants: { variant: 'default' },
});

export interface AlertProps
  extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof alertVariants> {}

const Alert: React.FC<AlertProps> = ({ className, variant, ...props }) => (
  <div role="alert" className={cn(alertVariants({ variant, className }))} {...props} />
);

export { Alert };
