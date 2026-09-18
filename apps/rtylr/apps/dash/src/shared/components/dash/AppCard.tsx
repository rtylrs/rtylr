import React from 'react';
import { useTranslation } from 'react-i18next';
import { ExternalLink, ArrowRight, Lock } from 'lucide-react';
import { motion } from 'motion/react';
import type { RtylrApp } from '../../hooks/queries/useApps';

export interface AppCardProps {
  app: RtylrApp;
  onLaunch?: (app: RtylrApp) => void;
}

export const AppCard: React.FC<AppCardProps> = ({ app, onLaunch }) => {
  const { t } = useTranslation();
  if (!app) return null;
  const Icon = app.icon;
  const isExternal = (app.href ?? '').startsWith('http');
  const isComingSoon = app.status === 'coming_soon';
  const isLocked = app.locked === true;

  const badge = (() => {
    if (isLocked) {
      return (
        <span
          className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase"
          style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
        >
          <Lock size={9} />
          {t('common.appCard.upgrade')}
        </span>
      );
    }
    if (app.status === 'beta') {
      return (
        <span
          className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase"
          style={{ background: 'rgba(139,92,246,0.15)', color: 'var(--app-crm)' }}
        >
          {t('common.appCard.beta')}
        </span>
      );
    }
    if (isComingSoon) {
      return (
        <span
          className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase"
          style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
        >
          {t('common.appCard.comingSoon')}
        </span>
      );
    }
    return isExternal ? (
      <ExternalLink size={13} style={{ color: 'var(--dash-text-dim)' }} />
    ) : (
      <ArrowRight size={13} style={{ color: 'var(--dash-text-dim)' }} />
    );
  })();

  const content = (
    <div className="flex min-w-0 items-start gap-3">
      <div
        className="h-11 w-11 shrink-0 rounded-xl flex items-center justify-center"
        style={{ background: app.bg, opacity: isLocked ? 0.5 : 1 }}
      >
        <Icon size={22} style={{ color: app.color }} />
      </div>
      <div className="min-w-0 flex-1">
        <p
          className="text-sm font-semibold leading-snug"
          style={{ color: isLocked ? 'var(--dash-text-muted)' : 'var(--dash-text)' }}
        >
          {app.label}
        </p>
        <p className="mt-1 text-xs leading-relaxed" style={{ color: 'var(--dash-text-muted)' }}>
          {app.description}
        </p>
      </div>
      <div className="flex shrink-0 items-center gap-1.5">{badge}</div>
    </div>
  );

  const cardStyle: React.CSSProperties = {
    background: 'var(--dash-card)',
    border: isLocked ? '1px solid var(--status-trial-bg)' : '1px solid var(--dash-border)',
    opacity: isComingSoon ? 0.6 : 1,
  };

  if (isComingSoon || (isLocked && false)) {
    return (
      <motion.div
        className="dash-app-card rounded-xl p-4 cursor-not-allowed"
        style={cardStyle}
        whileHover={{ scale: 1 }}
      >
        {content}
      </motion.div>
    );
  }

  // Locked apps link to the subscription upgrade page instead of launching the app
  if (isLocked) {
    return (
      <motion.a
        href="/settings/billing"
        className="dash-app-card group rounded-xl p-4 no-underline transition-all"
        style={cardStyle}
        whileHover={{ scale: 1.01 }}
      >
        {content}
      </motion.a>
    );
  }

  return (
    <motion.a
      href={app.href}
      target={isExternal ? '_blank' : undefined}
      rel={isExternal ? 'noopener noreferrer' : undefined}
      className="dash-app-card group rounded-xl p-4 no-underline transition-all"
      style={cardStyle}
      whileHover={{ scale: 1.01 }}
      onClick={() => onLaunch?.(app)}
    >
      {content}
    </motion.a>
  );
};

export default AppCard;
