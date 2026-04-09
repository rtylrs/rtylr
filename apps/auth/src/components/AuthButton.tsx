import React from 'react';
import { motion } from 'framer-motion';

export interface AuthButtonProps
  extends Omit<
    React.ButtonHTMLAttributes<HTMLButtonElement>,
    | 'onDrag'
    | 'onDragEnd'
    | 'onDragStart'
    | 'onAnimationStart'
    | 'onAnimationEnd'
    | 'onAnimationIteration'
  > {
  loading?: boolean;
  variant?: 'primary' | 'secondary' | 'outline';
}

/**
 * Primary auth action button. Solid brand blue, full-width, 52px tall,
 * with subtle press/hover animation via framer-motion.
 */
const AuthButton: React.FC<AuthButtonProps> = ({
  loading,
  variant = 'primary',
  disabled,
  children,
  className,
  ...props
}) => {
  const base =
    'relative w-full h-[52px] rounded-xl font-semibold text-[15px] transition-all duration-200 ' +
    'disabled:opacity-60 disabled:cursor-not-allowed ' +
    'focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#050617]';

  const variants: Record<NonNullable<AuthButtonProps['variant']>, string> = {
    primary:
      'bg-[#4583C4] text-white shadow-[0_8px_24px_-8px_rgba(69,131,196,0.6)] ' +
      'hover:bg-[#4e90d4] hover:shadow-[0_12px_32px_-8px_rgba(69,131,196,0.75)] ' +
      'active:bg-[#3e77b4] focus-visible:ring-[#4583C4]',
    secondary:
      'bg-[#64C5B9] text-[#05061a] ' +
      'hover:bg-[#72d1c5] active:bg-[#55b5a8] focus-visible:ring-[#64C5B9]',
    outline:
      'bg-white/[0.03] text-white border border-white/12 ' +
      'hover:bg-white/[0.08] hover:border-white/20 focus-visible:ring-white/30',
  };

  return (
    <motion.button
      whileHover={disabled || loading ? undefined : { scale: 1.005 }}
      whileTap={disabled || loading ? undefined : { scale: 0.985 }}
      transition={{ type: 'spring', stiffness: 500, damping: 30 }}
      disabled={disabled || loading}
      className={[base, variants[variant], className || ''].join(' ')}
      {...props}
    >
      {loading ? (
        <span className="inline-flex items-center justify-center gap-2">
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin-slow" />
          <span>Please wait…</span>
        </span>
      ) : (
        children
      )}
    </motion.button>
  );
};

export default AuthButton;
