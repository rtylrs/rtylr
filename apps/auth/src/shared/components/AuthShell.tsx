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
  subtitle?: string;
  /**
   * Optional footer slot rendered below the form (agreement text,
   * alternate-route prompts, etc.).
   */
  footer?: React.ReactNode;
}

/**
 * Split-screen authentication layout.
 *
 * - Left (60%): branded dark showcase with floating orbs (hidden on mobile).
 * - Right (40%): dark form panel with the actual auth UI.
 *
 * On mobile the showcase collapses and the right panel becomes full-width,
 * still on the same dark background for brand consistency.
 */
const AuthShell: React.FC<AuthShellProps> = ({ children, title, subtitle, footer }) => {
  return (
    <div className="min-h-screen w-full flex antialiased">
      <AuthShowcase />

      {/* Right: form panel - white background */}
      <main className="relative flex-1 lg:w-[40%] flex flex-col bg-white">
        {/* Mobile logo bar */}
        <div className="relative z-10 flex items-center justify-between px-6 pt-6 lg:hidden">
          <Logo variant="dark" className="h-7 w-auto select-none" />
        </div>

        {/* Centered form container */}
        <div className="relative z-10 flex-1 flex items-center justify-center px-6 py-10 sm:px-10 lg:px-14">
          <motion.div
            initial={{ opacity: 0, x: 24 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
            className="w-full max-w-md"
          >
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="mb-8"
            >
              <h2 className="text-3xl sm:text-[32px] font-semibold tracking-tight text-[#111]">
                {title}
              </h2>
              {subtitle && <p className="mt-2 text-[15px] text-[#666]">{subtitle}</p>}
            </motion.div>

            {children}

            {footer && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.5 }}
                className="mt-8"
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
