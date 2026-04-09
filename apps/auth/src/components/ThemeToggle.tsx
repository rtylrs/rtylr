import React from 'react';
import { Moon, Sun } from 'lucide-react';
import { motion } from 'framer-motion';

export interface ThemeToggleProps {
  theme?: 'light' | 'dark';
  onChange?: (theme: 'light' | 'dark') => void;
}

/**
 * Visual theme toggle. Placeholder only, no global theme wiring (the
 * auth app is locked to dark right now).
 */
const ThemeToggle: React.FC<ThemeToggleProps> = ({
  theme = 'dark',
  onChange,
}) => {
  return (
    <button
      type="button"
      aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
      onClick={() => onChange?.(theme === 'dark' ? 'light' : 'dark')}
      className="relative inline-flex h-8 w-14 items-center rounded-full border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] transition-colors"
    >
      <motion.div
        layout
        transition={{ type: 'spring', stiffness: 500, damping: 30 }}
        className={[
          'absolute top-0.5 flex h-7 w-7 items-center justify-center rounded-full',
          theme === 'dark' ? 'left-0.5 bg-[#4583C4]' : 'left-[26px] bg-amber-300',
        ].join(' ')}
      >
        {theme === 'dark' ? (
          <Moon className="h-3.5 w-3.5 text-white" />
        ) : (
          <Sun className="h-3.5 w-3.5 text-amber-900" />
        )}
      </motion.div>
    </button>
  );
};

export default ThemeToggle;
