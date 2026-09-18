import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { X, Clock, Zap } from 'lucide-react';
import { SubscriptionServiceClient } from '@voxire/voxie/hub/subscription/v1/client';
import { resolveAppUrl } from '@voxire/config';
import { readLocalStorage, writeLocalStorage } from '../lib/storage';
import { todayIso } from '../lib/dateUtils';

const subscriptionClient = createOneKitClient(SubscriptionServiceClient, 'crm');

function daysLeft(trialEndsAt: string | null): number {
  if (!trialEndsAt) return 0;
  return Math.max(0, Math.ceil((new Date(trialEndsAt).getTime() - Date.now()) / 86_400_000));
}

const DISMISSED_KEY = 'trial_banner_dismissed_date';

const TrialBanner: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const today = todayIso();
  const [dismissed, setDismissed] = useState(() => readLocalStorage(DISMISSED_KEY) === today);

  const { data } = useQuery({
    queryKey: ['subscription-banner'],
    queryFn: () => subscriptionClient.getMySubscription({}),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });

  if (!data || data.status !== 'trial' || dismissed) return null;

  const days = daysLeft(data.trialEndsAt ?? null);
  if (days > 7) return null;

  const dismiss = (): void => {
    writeLocalStorage(DISMISSED_KEY, today);
    setDismissed(true);
  };

  return (
    <div
      className="flex items-center gap-3 px-4 py-2.5 text-sm shrink-0"
      role="alert"
      style={{
        background: 'rgba(245,158,11,0.12)',
        borderBottom: '1px solid rgba(245,158,11,0.25)',
      }}
    >
      <Clock size={14} style={{ color: '#f59e0b', flexShrink: 0 }} />
      <p className="flex-1 font-medium" style={{ color: '#92400e' }}>
        {days === 0
          ? 'Your trial ends today.'
          : `Your trial ends in ${days} day${days !== 1 ? 's' : ''}.`}{' '}
        {t('crm.trialBanner.addPayment')}
      </p>
      <a
        href={`${resolveAppUrl('dash')}/settings/billing`}
        className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold shrink-0"
        style={{ background: '#f59e0b', color: '#fff' }}
      >
        <Zap size={11} />
        {t('crm.trialBanner.activatePlan')}
      </a>
      <button
        type="button"
        aria-label={t('license.license.dismiss')}
        onClick={dismiss}
        className="h-6 w-6 flex items-center justify-center rounded shrink-0 hover:bg-amber-100 transition-colors"
        style={{ color: '#92400e' }}
      >
        <X size={12} />
      </button>
    </div>
  );
};

export default TrialBanner;
