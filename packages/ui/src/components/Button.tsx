import React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../lib/cn';

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 select-none',
  {
    variants: {
      variant: {
        primary: 'bg-primary text-white hover:opacity-90 active:opacity-80',
        secondary: 'bg-secondary text-white hover:opacity-90 active:opacity-80',
        accent: 'bg-accent text-white hover:opacity-90 active:opacity-80',
        outline: 'border border-current bg-transparent hover:bg-black/5 active:bg-black/10',
        ghost: 'bg-transparent hover:bg-black/5 active:bg-black/8',
        soft: 'bg-primary/10 text-primary hover:bg-primary/15',
        danger: 'bg-destructive text-white hover:opacity-90',
        'dark-outline': 'border border-white/20 text-white bg-transparent hover:bg-white/8',
      },
      size: {
        sm: 'h-8 px-3 text-xs',
        md: 'h-10 px-4 text-sm',
        lg: 'h-12 px-6 text-base',
      },
      fullWidth: {
        true: 'w-full',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
);

export interface ButtonProps
  extends
    Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'prefix'>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  loading?: boolean;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant,
      size,
      fullWidth,
      asChild,
      loading,
      prefix,
      suffix,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    const Comp = asChild ? Slot : 'button';
    return (
      <Comp
        ref={ref}
        className={cn(buttonVariants({ variant, size, fullWidth, className }))}
        disabled={disabled || loading}
        {...props}
      >
        {loading ? (
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin-slow" />
        ) : (
          prefix && <span className="shrink-0">{prefix}</span>
        )}
        {children}
        {!loading && suffix && <span className="shrink-0">{suffix}</span>}
      </Comp>
    );
  }
);

Button.displayName = 'Button';

export { Button, buttonVariants };
