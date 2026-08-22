import React from 'react';
import { cn } from '../../lib/utils';

type ButtonVariant = 'default' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'accent';
type ButtonSize = 'sm' | 'md' | 'lg' | 'icon';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  children?: React.ReactNode;
}

const variantClasses: Record<ButtonVariant, string> = {
  default: 'text-white',
  secondary: 'text-white/70 hover:text-white',
  outline: 'text-white/60 hover:text-white/90',
  ghost: 'text-white/50 hover:text-white/80',
  destructive: 'bg-red-600/90 text-white hover:bg-red-600',
  accent: 'text-white hover:opacity-90',
};

const variantStyles: Record<ButtonVariant, React.CSSProperties> = {
  default: { background: 'var(--accent)' },
  secondary: { background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' },
  outline: {
    background: 'transparent',
    border: '1px solid var(--accent-border)',
    color: 'var(--accent)',
  },
  ghost: { background: 'transparent' },
  destructive: {},
  accent: { background: 'var(--accent)' },
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'h-7 px-3 text-xs rounded-md',
  md: 'h-9 px-4 text-sm rounded-lg',
  lg: 'h-11 px-6 text-sm rounded-lg',
  icon: 'h-9 w-9 rounded-lg flex items-center justify-center',
};

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'default', size = 'md', className, children, ...props }, ref) => {
    return (
      <button
        type="button"
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center gap-2 font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]/40 disabled:pointer-events-none disabled:opacity-40 hover:opacity-90',
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        style={{ ...variantStyles[variant], ...props.style }}
        {...props}
      >
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };
export type { ButtonProps, ButtonVariant, ButtonSize };
