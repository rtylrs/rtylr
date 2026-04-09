import React, { useState, forwardRef } from 'react';
import { Eye, EyeOff, LucideIcon } from 'lucide-react';

export interface AuthInputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  icon?: LucideIcon;
  error?: string;
}

/**
 * Dark-themed input with uppercase label, optional lucide icon prefix,
 * and an eye-toggle for password fields. Designed to match the rtylr
 * auth shell (dark bg, rounded-xl, teal focus ring).
 */
const AuthInput = forwardRef<HTMLInputElement, AuthInputProps>(
  ({ label, icon: Icon, error, id, type, className, ...props }, ref) => {
    const [showPassword, setShowPassword] = useState(false);
    const isPassword = type === 'password';
    const inputType = isPassword ? (showPassword ? 'text' : 'password') : type;

    return (
      <div className="flex flex-col gap-2">
        <label
          htmlFor={id}
          className="text-[11px] font-semibold uppercase tracking-[0.08em] text-white/50"
        >
          {label}
        </label>
        <div className="relative group">
          {Icon && (
            <Icon
              className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 h-[18px] w-[18px] text-white/35 group-focus-within:text-[#64C5B9] transition-colors"
              aria-hidden="true"
            />
          )}
          <input
            ref={ref}
            id={id}
            type={inputType}
            className={[
              'w-full h-[52px] rounded-xl bg-white/[0.04] text-[15px] text-white placeholder:text-white/30',
              'border border-white/10 outline-none transition-all duration-150',
              'hover:bg-white/[0.06] hover:border-white/15',
              'focus:bg-white/[0.06] focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/25',
              Icon ? 'pl-12' : 'pl-4',
              isPassword ? 'pr-12' : 'pr-4',
              error ? 'border-red-400/60 focus:border-red-400 focus:ring-red-400/25' : '',
              className || '',
            ].join(' ')}
            {...props}
          />
          {isPassword && (
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              className="absolute right-3 top-1/2 -translate-y-1/2 p-1.5 rounded-md text-white/40 hover:text-white/80 hover:bg-white/5 transition-colors focus:outline-none focus:ring-2 focus:ring-[#4583C4]/40"
            >
              {showPassword ? (
                <EyeOff className="h-[18px] w-[18px]" />
              ) : (
                <Eye className="h-[18px] w-[18px]" />
              )}
            </button>
          )}
        </div>
        {error && (
          <p className="text-xs text-red-300/90">{error}</p>
        )}
      </div>
    );
  }
);

AuthInput.displayName = 'AuthInput';

export default AuthInput;
