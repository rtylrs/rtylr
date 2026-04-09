import React from 'react';
import { ExternalLink, ArrowRight } from 'lucide-react';
import { motion } from 'framer-motion';
import type { RtylrApp } from '../../hooks/queries/useApps';

export interface AppCardProps {
  app: RtylrApp;
  onLaunch?: (app: RtylrApp) => void;
}

export const AppCard: React.FC<AppCardProps> = ({ app, onLaunch }) => {
  const Icon = app.icon;
  const isExternal = app.href.startsWith('http');
  const isComingSoon = app.status === 'coming_soon';

  const content = (
    <>
      <div className="flex items-start justify-between">
        <div className="h-11 w-11 rounded-xl flex items-center justify-center" style={{ background: app.bg }}>
          <Icon size={22} style={{ color: app.color }} />
        </div>
        <div className="flex items-center gap-1.5">
          {app.status === 'beta' && (
            <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase" style={{ background: 'rgba(139,92,246,0.15)', color: '#8b5cf6' }}>
              Beta
            </span>
          )}
          {isComingSoon && (
            <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}>
              Soon
            </span>
          )}
          {!isComingSoon && (isExternal ? <ExternalLink size={13} style={{ color: 'var(--hub-text-dim)' }} /> : <ArrowRight size={13} style={{ color: 'var(--hub-text-dim)' }} />)}
        </div>
      </div>
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{app.label}</p>
        <p className="text-xs mt-1 leading-relaxed" style={{ color: 'var(--hub-text-muted)' }}>
          {app.description}
        </p>
      </div>
    </>
  );

  const style: React.CSSProperties = {
    background: 'var(--hub-card)',
    border: '1px solid var(--hub-border)',
    opacity: isComingSoon ? 0.6 : 1,
  };

  if (isComingSoon) {
    return (
      <motion.div
        className="rounded-xl p-5 flex flex-col gap-4 cursor-not-allowed"
        style={style}
        whileHover={{ scale: 1 }}
      >
        {content}
      </motion.div>
    );
  }

  return (
    <motion.a
      href={app.href}
      target={isExternal ? '_blank' : undefined}
      rel={isExternal ? 'noopener noreferrer' : undefined}
      className="group rounded-xl p-5 flex flex-col gap-4 no-underline transition-all"
      style={style}
      whileHover={{ scale: 1.01 }}
      onClick={() => onLaunch?.(app)}
    >
      {content}
    </motion.a>
  );
};

export default AppCard;
