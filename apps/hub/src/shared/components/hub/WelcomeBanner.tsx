import React from 'react';
import { X, Sparkles } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

const STORAGE_KEY = 'hub.welcome.dismissed';

export interface WelcomeBannerProps {
  title?: string;
  description?: string;
  ctaLabel?: string;
  onCta?: () => void;
}

export const WelcomeBanner: React.FC<WelcomeBannerProps> = ({
  title = 'Welcome to rtylr Hub',
  description = 'Your universal command center. Launch apps, invite your team, and track everything in one place.',
  ctaLabel = 'Take a tour',
  onCta,
}) => {
  const [dismissed, setDismissed] = React.useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return localStorage.getItem(STORAGE_KEY) === '1';
  });

  if (dismissed) return null;

  const handleDismiss = () => {
    localStorage.setItem(STORAGE_KEY, '1');
    setDismissed(true);
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        className="relative rounded-xl p-5 overflow-hidden"
        style={{
          background: 'linear-gradient(135deg, rgba(69,131,196,0.14), rgba(100,197,185,0.1))',
          border: '1px solid rgba(69,131,196,0.3)',
        }}
      >
        <div className="flex items-start gap-4">
          <div
            className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
            style={{ background: 'rgba(69,131,196,0.2)' }}
          >
            <Sparkles size={18} style={{ color: '#4583C4' }} />
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{title}</p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-muted)' }}>{description}</p>
            {onCta && (
              <button
                onClick={onCta}
                className="mt-3 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-opacity hover:opacity-90"
                style={{ background: '#4583C4', color: 'white' }}
              >
                {ctaLabel}
              </button>
            )}
          </div>
          <button
            onClick={handleDismiss}
            className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-white/10 shrink-0"
            aria-label="Dismiss welcome"
          >
            <X size={14} style={{ color: 'var(--hub-text-dim)' }} />
          </button>
        </div>
      </motion.div>
    </AnimatePresence>
  );
};

export default WelcomeBanner;
