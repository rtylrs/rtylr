import React from 'react';
import { motion } from 'framer-motion';

export interface AuthCardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  className?: string;
  compact?: boolean;
}

/**
 * Reusable card wrapper. Uses the existing dark style with a subtle
 * glass surface, used by secondary screens like Sessions, Devices, etc.
 */
const AuthCard: React.FC<AuthCardProps> = ({
  children,
  title,
  subtitle,
  className = '',
  compact = false,
}) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
      className={[
        'rounded-2xl border border-white/10 bg-white/[0.03] backdrop-blur-sm',
        compact ? 'p-4' : 'p-5 sm:p-6',
        className,
      ].join(' ')}
    >
      {(title || subtitle) && (
        <div className="mb-4">
          {title && (
            <h3 className="text-[15px] font-semibold text-white">{title}</h3>
          )}
          {subtitle && (
            <p className="mt-1 text-[13px] text-white/55">{subtitle}</p>
          )}
        </div>
      )}
      {children}
    </motion.div>
  );
};

export default AuthCard;
