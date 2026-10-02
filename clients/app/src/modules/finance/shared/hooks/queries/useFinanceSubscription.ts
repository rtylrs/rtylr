import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { SubscriptionServiceClient } from '@voxire/voxie/hub/subscription/v1/client';
import type { Subscription, BillingDetail } from '@voxire/voxie/hub/subscription/v1/types';

// ── Types ─────────────────────────────────────────────────────────────────────

export type FinanceSubscription = Subscription;
export type FinanceBillingDetail = BillingDetail;
const subscriptionClient = createOneKitClient(SubscriptionServiceClient, 'finance');

// ── Helpers ───────────────────────────────────────────────────────────────────

export function centsToDisplay(cents: number): string {
  const safeCents = Number.isFinite(cents) ? cents : 0;
  return `$${(safeCents / 100).toFixed(2)}`;
}

export function trialDaysLeft(trialEndsAt: string | null): number {
  if (!trialEndsAt) return 0;
  return Math.max(0, Math.ceil((new Date(trialEndsAt).getTime() - Date.now()) / 86_400_000));
}

export interface SeatProration {
  daysRemaining: number;
  daysInMonth: number;
  proratedCents: number;
  nextMonthName: string;
}

export function computeSeatProration(pricePerSeatCents: number): SeatProration {
  const now = new Date();
  const month = now.getMonth();
  const daysInMonth = new Date(now.getFullYear(), month + 1, 0).getDate();
  const daysRemaining = daysInMonth - now.getDate() + 1;
  const MONTHS = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  return {
    daysRemaining,
    daysInMonth,
    proratedCents: Math.round((daysRemaining * pricePerSeatCents) / daysInMonth),
    nextMonthName: MONTHS[(month + 1) % 12],
  };
}

export function formatPeriodRange(start: string, end: string | null): string {
  const s = new Date(start);
  const e = end ? new Date(end) : new Date();
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  const sMonth = months[s.getUTCMonth()];
  const eMonth = months[e.getUTCMonth()];
  const year = e.getUTCFullYear();
  if (s.getUTCMonth() === e.getUTCMonth() && s.getUTCFullYear() === e.getUTCFullYear()) {
    return `${sMonth} ${s.getUTCDate()} – ${e.getUTCDate()}, ${year}`;
  }
  return `${sMonth} ${s.getUTCDate()} – ${eMonth} ${e.getUTCDate()}, ${year}`;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useFinanceSubscription = (): ReturnType<
  typeof useQuery<FinanceSubscription | undefined>
> =>
  useQuery({
    queryKey: ['finance', 'finance-subscription'],
    queryFn: () => subscriptionClient.getMySubscription({}),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });

export const useFinanceSubscriptionBilling = (): ReturnType<
  typeof useQuery<FinanceBillingDetail | undefined>
> =>
  useQuery({
    queryKey: ['finance', 'finance-subscription', 'billing'],
    queryFn: () => subscriptionClient.getBillingDetail({}),
    select: (d) => d.billing,
    staleTime: 2 * 60 * 1000,
  });
