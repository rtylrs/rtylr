import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface FlowSubscription {
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

export interface FlowBillingLineItem {
  periodStart: string;
  periodEnd: string | null;
  seatCount: number;
  pricePerSeatCents: number;
  amountCents: number;
}

export interface FlowBillingDetail {
  billingMonth: string;
  nextBillingDate: string | null;
  lineItems: FlowBillingLineItem[];
  totalCents: number;
  planName: string;
  status: 'active' | 'trial' | 'cancelled' | 'expired';
  trialEndsAt: string | null;
  paddleSubscriptionId: string | null;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

export function centsToDisplay(cents: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en', { style: 'currency', currency }).format(cents / 100);
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

function shortMonth(date: Date, utc = false): string {
  return new Intl.DateTimeFormat('en', {
    month: 'short',
    timeZone: utc ? 'UTC' : undefined,
  }).format(date);
}

export function computeSeatProration(pricePerSeatCents: number): SeatProration {
  const now = new Date();
  const month = now.getMonth();
  const daysInMonth = new Date(now.getFullYear(), month + 1, 0).getDate();
  const daysRemaining = daysInMonth - now.getDate() + 1;
  const nextMonthDate = new Date(now.getFullYear(), (month + 1) % 12, 1);
  return {
    daysRemaining,
    daysInMonth,
    proratedCents: Math.round((daysRemaining * pricePerSeatCents) / daysInMonth),
    nextMonthName: shortMonth(nextMonthDate),
  };
}

export function formatPeriodRange(start: string, end: string | null): string {
  const s = new Date(start);
  const e = end ? new Date(end) : new Date();
  const sMonth = shortMonth(s, true);
  const eMonth = shortMonth(e, true);
  const year = e.getUTCFullYear();
  if (s.getUTCMonth() === e.getUTCMonth() && s.getUTCFullYear() === e.getUTCFullYear()) {
    return `${sMonth} ${s.getUTCDate()} – ${e.getUTCDate()}, ${year}`;
  }
  return `${sMonth} ${s.getUTCDate()} – ${eMonth} ${e.getUTCDate()}, ${year}`;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

interface SubResponse {
  subscription: FlowSubscription | null;
}

interface BillingResponse {
  billing: FlowBillingDetail | null;
}

export const useFlowSubscription = (): ReturnType<typeof useQuery<FlowSubscription | null>> =>
  useQuery({
    queryKey: ['flow-subscription'],
    queryFn: () => unwrap(safe.get<SubResponse>('/subscriptions/my')),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });

export const useFlowBillingDetail = (): ReturnType<typeof useQuery<FlowBillingDetail | null>> =>
  useQuery({
    queryKey: ['flow-subscription', 'billing'],
    queryFn: () => unwrap(safe.get<BillingResponse>('/subscriptions/billing')),
    select: (d) => d.billing,
    staleTime: 2 * 60 * 1000,
  });
