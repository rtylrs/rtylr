import React from 'react';
import { motion } from 'motion/react';
import AuthShowcase from './AuthShowcase';
import Logo from './Logo';

interface AuthShellProps {
  children: React.ReactNode;
  /**
   * Page-specific heading shown above the form on the right side.
   */
  title: string;
  /**
   * Page-specific subtitle shown below the title.
   */
  subtitle?: React.ReactNode;
  /**
   * Optional footer slot rendered below the form (agreement text,
   * alternate-route prompts, etc.).
   */
  footer?: React.ReactNode;
}

/**
 * Split-screen authentication layout, in the marketing site's light style.
 *
 * - Left (52%): the rtylr pitch and a few product cards (hidden below lg).
 * - Right: the form on white, centred, max 400px.
 *
 * Every form screen (sign in, sign up, reset, invite, consent) renders
 * through this, so they stay one product instead of five.
 */
const AuthShell: React.FC<AuthShellProps> = ({ children, title, subtitle, footer }) => {
  return (
    <div className="min-h-screen w-full flex antialiased bg-white">
      <AuthShowcase />

      <main className="relative flex-1 flex flex-col bg-white">
        {/* Mobile logo bar */}
        <div className="relative z-10 flex items-center px-6 pt-6 lg:hidden">
          <Logo variant="dark" className="h-7 w-auto select-none" />
        </div>

        <div className="relative z-10 flex-1 flex items-center justify-center px-6 py-10 sm:px-10 lg:px-16">
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
            className="w-full max-w-[400px]"
          >
            <div className="mb-8">
              <h2 className="text-[28px] font-semibold tracking-[-0.03em] text-[#111]">{title}</h2>
              {subtitle && (
                <p className="mt-2 text-[15px] leading-relaxed text-[#111]/55">{subtitle}</p>
              )}
            </div>

            {children}

            {footer && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.4, delay: 0.25 }}
                className="mt-8 text-sm text-[#111]/55"
              >
                {footer}
              </motion.div>
            )}
          </motion.div>
        </div>
      </main>
    </div>
  );
};

export default AuthShell;
