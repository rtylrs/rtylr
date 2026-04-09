import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { AlertTriangle } from 'lucide-react';

export interface RateLimitBannerProps {
  seconds: number;
  onExpire?: () => void;
}

const RateLimitBanner: React.FC<RateLimitBannerProps> = ({
  seconds,
  onExpire,
}) => {
  const [remaining, setRemaining] = useState(seconds);

  useEffect(() => {
    setRemaining(seconds);
  }, [seconds]);

  useEffect(() => {
    if (remaining <= 0) {
      onExpire?.();
      return;
    }
    const id = setTimeout(() => setRemaining((r) => r - 1), 1000);
    return () => clearTimeout(id);
  }, [remaining, onExpire]);

  const mm = Math.floor(remaining / 60);
  const ss = remaining % 60;
  const label =
    mm > 0
      ? `${mm}:${ss.toString().padStart(2, '0')}`
      : `${ss} second${ss === 1 ? '' : 's'}`;

  if (remaining <= 0) return null;

  return (
    <motion.div
      initial={{ opacity: 0, y: -6 }}
      animate={{ opacity: 1, y: 0 }}
      role="alert"
      className="flex items-start gap-2 rounded-lg border border-amber-400/25 bg-amber-500/10 px-3.5 py-2.5"
    >
      <AlertTriangle className="h-4 w-4 text-amber-300 mt-0.5 shrink-0" />
      <p className="text-xs text-amber-100 leading-relaxed">
        Too many attempts. Try again in <strong className="tabular-nums">{label}</strong>.
      </p>
    </motion.div>
  );
};

export default RateLimitBanner;
