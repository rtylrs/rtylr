import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { useUserStore } from '@voxire/auth';
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

const DISMISSED_KEY_PREFIX = 'trial_banner_dismissed_date';

function nonEmptyString(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value : null;
}

const TrialBanner: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const [today, setToday] = useState(todayIso);
  const [dismissedScope, setDismissedScope] = useState<string | null>(null);
  const [dismissedDate, setDismissedDate] = useState<string | null>(null);

  useEffect(() => {
    let timeout: number | undefined;
    const refresh = (): void => setToday(todayIso());
    const schedule = (): void => {
      const now = new Date();
      const nextMidnight = new Date(now);
      nextMidnight.setHours(24, 0, 0, 0);
      timeout = window.setTimeout(
        () => {
          refresh();
          schedule();
        },
        Math.max(1000, nextMidnight.getTime() - now.getTime())
      );
    };

    schedule();
    window.addEventListener('focus', refresh);
    document.addEventListener('visibilitychange', refresh);
    return () => {
      if (timeout !== undefined) window.clearTimeout(timeout);
      window.removeEventListener('focus', refresh);
      document.removeEventListener('visibilitychange', refresh);
    };
  }, []);

  const userId = nonEmptyString(user?.id) ?? nonEmptyString(user?.userId);
  const businessId =
    nonEmptyString(user?.effectiveAccess?.businessId) ?? nonEmptyString(user?.businessId);
  const dismissalScope = userId
    ? `${DISMISSED_KEY_PREFIX}:${userId}:${businessId ?? 'business'}`
    : null;
  const dismissed =
    dismissalScope !== null &&
    ((dismissedScope === dismissalScope && dismissedDate === today) ||
      readLocalStorage(dismissalScope) === today);

  const { data } = useQuery({
    queryKey: ['subscription-banner', dismissalScope],
    queryFn: () => subscriptionClient.getMySubscription({}),
    select: (d) => d.subscription,
    enabled: dismissalScope !== null,
    staleTime: 5 * 60 * 1000,
  });

  if (dismissalScope === null || !data || data.status !== 'trial' || dismissed) return null;

  const days = daysLeft(data.trialEndsAt ?? null);
  if (days > 7) return null;

  const dismiss = (): void => {
    writeLocalStorage(dismissalScope, today);
    setDismissedScope(dismissalScope);
    setDismissedDate(today);
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
        aria-label={t('license.dismiss')}
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
