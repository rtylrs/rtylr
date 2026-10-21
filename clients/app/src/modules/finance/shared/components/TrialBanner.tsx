import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { X, Clock, Zap } from 'lucide-react';
import { resolveAppUrl } from '@voxire/config';
import { useFinanceSubscription } from '../hooks/queries/useFinanceSubscription';
import { todayIso } from '../lib/date';

function daysLeft(trialEndsAt: string | null): number {
  if (!trialEndsAt) return 0;
  return Math.max(0, Math.ceil((new Date(trialEndsAt).getTime() - Date.now()) / 86_400_000));
}

const DISMISSED_KEY = 'trial_banner_dismissed_date';

const TrialBanner: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation('financeCommon');
  const today = todayIso();
  const [dismissed, setDismissed] = useState(() => localStorage.getItem(DISMISSED_KEY) === today);

  const { data } = useFinanceSubscription();

  if (!data || data.status !== 'trial' || dismissed) return null;

  const days = daysLeft(data.trialEndsAt ?? null);
  if (days > 7) return null;

  const dismiss = (): void => {
    localStorage.setItem(DISMISSED_KEY, today);
    setDismissed(true);
  };

  return (
    <div
      className="flex items-center gap-3 px-4 py-2.5 text-sm shrink-0"
      role="alert"
      style={{
        background: 'var(--finance-warning-soft)',
        borderBottom: '1px solid var(--finance-warning-border)',
      }}
    >
      <Clock size={14} style={{ color: 'var(--finance-warning)', flexShrink: 0 }} />
      <p className="flex-1 font-medium" style={{ color: 'var(--finance-warning-text)' }}>
        {days === 0
          ? 'Your trial ends today.'
          : `Your trial ends in ${days} day${days !== 1 ? 's' : ''}.`}{' '}
        {t('finance.trial.addPayment')}
      </p>
      <a
        href={`${resolveAppUrl('dash')}/settings/billing`}
        className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold shrink-0"
        style={{ background: 'var(--finance-warning)', color: 'var(--finance-white)' }}
      >
        <Zap size={11} />
        {t('finance.trial.activate')}
      </a>
      <button
        type="button"
        aria-label={t('finance.trial.dismiss')}
        onClick={dismiss}
        className="h-6 w-6 flex items-center justify-center rounded shrink-0 hover:bg-amber-100 transition-colors"
        style={{ color: 'var(--finance-warning-text)' }}
      >
        <X size={12} />
      </button>
    </div>
  );
};

export default TrialBanner;
