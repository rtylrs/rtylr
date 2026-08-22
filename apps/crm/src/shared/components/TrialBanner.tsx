import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { X, Clock, Zap } from 'lucide-react';
import { SubscriptionServiceClient } from '@voxire/voxie/hub/subscription/v1/client';
import { resolveAppUrl } from '@voxire/config';
import { trialDaysLeft } from '../hooks/queries/useCrmSubscription';

const subscriptionClient = createOneKitClient(SubscriptionServiceClient, 'crm');

const DISMISSED_KEY = 'trial_banner_dismissed_date';

function localDateKey(date = new Date()): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

const TrialBanner: React.FC = (): React.JSX.Element | null => {
  const today = localDateKey();
  const [dismissed, setDismissed] = useState(() => localStorage.getItem(DISMISSED_KEY) === today);

  const { data } = useQuery({
    queryKey: ['subscription-banner'],
    queryFn: () => subscriptionClient.getMySubscription({}),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });

  if (!data || data.status !== 'trial' || dismissed) return null;

  const days = trialDaysLeft(data.trialEndsAt ?? null);
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
        background: 'rgba(245,158,11,0.12)',
        borderBottom: '1px solid rgba(245,158,11,0.25)',
      }}
    >
      <Clock size={14} style={{ color: '#f59e0b', flexShrink: 0 }} />
      <p className="flex-1 font-medium" style={{ color: '#92400e' }}>
        {days === 0
          ? 'Your trial ends today.'
          : `Your trial ends in ${days} day${days !== 1 ? 's' : ''}.`}{' '}
        Add a payment method to keep your data.
      </p>
      <a
        href={`${resolveAppUrl('dash')}/settings/billing`}
        className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold shrink-0"
        style={{ background: '#f59e0b', color: '#fff' }}
      >
        <Zap size={11} />
        Activate plan
      </a>
      <button
        type="button"
        aria-label="Dismiss"
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
