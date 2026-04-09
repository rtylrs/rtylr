import React from 'react';
import { motion } from 'framer-motion';
import { FaGoogle, FaApple, FaMicrosoft, FaGithub } from 'react-icons/fa';

export type SocialProvider = 'google' | 'apple' | 'microsoft' | 'github';

export interface SocialLoginButtonProps {
  provider: SocialProvider;
  onClick?: () => void;
  disabled?: boolean;
  label?: string;
}

const PROVIDER_META: Record<
  SocialProvider,
  { icon: React.ComponentType<{ className?: string }>; label: string }
> = {
  google: { icon: FaGoogle, label: 'Continue with Google' },
  apple: { icon: FaApple, label: 'Continue with Apple' },
  microsoft: { icon: FaMicrosoft, label: 'Continue with Microsoft' },
  github: { icon: FaGithub, label: 'Continue with GitHub' },
};

const SocialLoginButton: React.FC<SocialLoginButtonProps> = ({
  provider,
  onClick,
  disabled,
  label,
}) => {
  const meta = PROVIDER_META[provider];
  const Icon = meta.icon;
  return (
    <motion.button
      type="button"
      whileHover={disabled ? undefined : { scale: 1.005 }}
      whileTap={disabled ? undefined : { scale: 0.985 }}
      onClick={onClick}
      disabled={disabled}
      aria-label={label || meta.label}
      className={[
        'w-full h-[52px] rounded-xl bg-white/[0.03] border border-white/12 text-white',
        'text-[15px] font-semibold transition-all duration-200',
        'hover:bg-white/[0.08] hover:border-white/20',
        'focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30',
        'disabled:opacity-60 disabled:cursor-not-allowed',
      ].join(' ')}
    >
      <span className="inline-flex items-center justify-center gap-2.5">
        <Icon className="text-white/85" />
        {label || meta.label}
      </span>
    </motion.button>
  );
};

export default SocialLoginButton;
