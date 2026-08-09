import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { SubscriptionServiceClient } from '@voxire/voxie/hub/subscription/v1/client';

const subscriptionClient = createOneKitClient(SubscriptionServiceClient, 'flow');

// ── Types ─────────────────────────────────────────────────────────────────────

export interface FlowSubscription {
  id: string;
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

const VALID_STATUSES: FlowSubscription['status'][] = ['active', 'trial', 'cancelled', 'expired'];

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useFlowSubscription = (): ReturnType<typeof useQuery<FlowSubscription | null>> =>
  useQuery({
    queryKey: ['flow-subscription'],
    queryFn: async (): Promise<FlowSubscription | null> => {
      const res = await subscriptionClient.getMySubscription({});
      const sub = res.subscription;
      if (!sub) return null;
      return {
        id: sub.id ?? '',
        planName: sub.planName ?? '',
        planSlug: sub.planSlug ?? '',
        status:
          sub.status && (VALID_STATUSES as string[]).includes(sub.status)
            ? (sub.status as FlowSubscription['status'])
            : 'active',
        seatCount: sub.seatCount ?? 0,
        pricePerSeatCents: sub.pricePerSeatCents ?? 0,
        trialEndsAt: sub.trialEndsAt ?? null,
        nextBillingDate: sub.nextBillingDate ?? null,
        paddleSubscriptionId: sub.paddleSubscriptionId || null,
      };
    },
    staleTime: 5 * 60 * 1000,
  });
