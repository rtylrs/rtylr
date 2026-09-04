import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { X, Clock, Zap } from 'lucide-react';
import { resolveAppUrl } from '@voxire/config';
import { useFlowSubscription } from '../hooks/queries/useFlowSubscription';

function daysLeft(trialEndsAt: string | null): number {
  if (!trialEndsAt) return 0;
  return Math.max(0, Math.ceil((new Date(trialEndsAt).getTime() - Date.now()) / 86_400_000));
}

const DISMISSED_KEY = 'trial_banner_dismissed';

const TrialBanner: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const [dismissed, setDismissed] = useState(() => localStorage.getItem(DISMISSED_KEY) === '1');

  const { data } = useFlowSubscription();

  if (!data || data.status !== 'trial' || dismissed) return null;

  const days = daysLeft(data.trialEndsAt);
  if (days > 7) return null;

  const dismiss = (): void => {
    localStorage.setItem(DISMISSED_KEY, '1');
    setDismissed(true);
  };

  return (
    <div
      className="flex items-center gap-3 px-4 py-2.5 text-sm shrink-0"
      role="alert"
      style={{
        // The pending chip family, which is Flow's amber notice tone and is
        // contrast-checked per mode. The literals here were the same amber by
        // eye but did not change with the theme.
        background: 'var(--flow-status-pending-bg)',
        borderBottom: '1px solid var(--flow-status-pending-border)',
      }}
    >
      <Clock size={14} style={{ color: 'var(--flow-warn-fg)', flexShrink: 0 }} />
      <p className="flex-1 font-medium" style={{ color: 'var(--flow-status-pending)' }}>
        {/*
          The day count goes through i18next rather than a `days !== 1 ? 's'`
          suffix, which only ever worked for English.
        */}
        {days === 0
          ? t('flow.trialBanner.endsToday')
          : t('flow.trialBanner.endsIn', { count: days })}{' '}
        {t('flow.trialBanner.addPayment')}
      </p>
      <a
        href={`${resolveAppUrl('dash')}/settings/billing`}
        className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold shrink-0"
        style={{ background: 'var(--flow-warn-fg)', color: '#fff' }}
      >
        <Zap size={11} />
        {t('flow.trialBanner.activate')}
      </a>
      <button
        type="button"
        aria-label={t('flow.trialBanner.dismiss')}
        onClick={dismiss}
        className="h-6 w-6 flex items-center justify-center rounded-sm shrink-0 hover:bg-amber-100 transition-colors"
        style={{ color: 'var(--flow-status-pending)' }}
      >
        <X size={12} />
      </button>
    </div>
  );
};

export default TrialBanner;
