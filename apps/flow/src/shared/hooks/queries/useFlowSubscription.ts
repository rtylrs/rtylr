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

// ── Helpers ───────────────────────────────────────────────────────────────────

export function centsToDisplay(cents: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en', { style: 'currency', currency }).format(cents / 100);
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

// ── Hooks ─────────────────────────────────────────────────────────────────────

interface SubResponse {
  subscription: FlowSubscription | null;
}

export const useFlowSubscription = (): ReturnType<typeof useQuery<FlowSubscription | null>> =>
  useQuery({
    queryKey: ['flow-subscription'],
    queryFn: () => unwrap(safe.get<SubResponse>('/subscriptions/my')),
    select: (d) => d.subscription,
    staleTime: 5 * 60 * 1000,
  });
