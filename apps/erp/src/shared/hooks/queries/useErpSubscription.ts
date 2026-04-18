import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ErpSubscription {
  id: number;
  planName: string;
  planSlug: string;
  status: 'active' | 'trial' | 'cancelled' | 'expired';
  seatCount: number;
  pricePerSeatCents: number;
  trialEndsAt: string | null;
  nextBillingDate: string | null;
  paddleSubscriptionId: string | null;
}

export interface ErpBillingLineItem {
  periodStart: string;
  periodEnd: string | null;
  seatCount: number;
  pricePerSeatCents: number;
  amountCents: number;
}

export interface ErpBillingDetail {
  billingMonth: string;
  nextBillingDate: string | null;
  lineItems: ErpBillingLineItem[];
  totalCents: number;
  planName: string;
  status: 'active' | 'trial' | 'cancelled' | 'expired';
  trialEndsAt: string | null;
  paddleSubscriptionId: string | null;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

export function centsToDisplay(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
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

interface SubResponse {
  subscription: ErpSubscription | null;
}

interface BillingResponse {
  billing: ErpBillingDetail | null;
}

export const useErpSubscription = (): ReturnType<typeof useQuery<ErpSubscription | null>> =>
  useQuery({
    queryKey: ['erp-subscription'],
    queryFn: () => unwrap(safe.get<SubResponse>('/subscriptions/my')),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });

export const useErpSubscriptionBilling = (): ReturnType<typeof useQuery<ErpBillingDetail | null>> =>
  useQuery({
    queryKey: ['erp-subscription', 'billing'],
    queryFn: () => unwrap(safe.get<BillingResponse>('/subscriptions/billing')),
    select: (d) => d.billing,
    staleTime: 2 * 60 * 1000,
  });
