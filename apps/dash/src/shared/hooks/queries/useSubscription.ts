import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { safe } from '../../lib/api';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface SubscriptionPayload {
  id: number;
  planName: string;
  planSlug: string;
  apps: string[];
  pricePerSeatCents: number;
  status: 'active' | 'trial' | 'cancelled' | 'expired';
  isGranted?: boolean;
  seatCount: number;
  includedSeatLimit: number;
  includedSeatCount: number;
  billableSeatCount: number;
  grantExpiresAt?: string | null;
  trialEndsAt: string | null;
  nextBillingDate: string | null;
  paddleSubscriptionId: string | null;
}

export interface PlanPayload {
  id: number;
  name: string;
  slug: string;
  apps: string[];
  pricePerSeatCents: number;
  annualPricePerSeatCents: number;
  trialDays: number;
  tagline: string;
  maxSeats: number | null;
  maxLocations: number | null;
  aiCreditsPerMonth: number;
  features: string[];
  tierOrder: number;
  isSolo: boolean;
  isFeatured: boolean;
  isActive: boolean;
}

export interface BillingLineItem {
  periodStart: string;
  periodEnd: string | null; // null = still active
  seatCount: number;
  includedSeatLimit?: number;
  includedSeatCount?: number;
  billableSeatCount?: number;
  includedCents?: number;
  pricePerSeatCents: number;
  amountCents: number;
  type?: 'seat' | 'addon';
  description?: string;
  isGranted?: boolean;
  grantExpiresAt?: string | null;
  addonId?: string;
  quantity?: number;
  unitPriceCents?: number;
}

export interface BillingDetailPayload {
  billingMonth: string;
  nextBillingDate: string | null;
  lineItems: BillingLineItem[];
  totalCents: number;
  planName: string;
  pricePerSeatCents: number;
  currentSeatCount: number;
  includedSeatLimit: number;
  includedSeatCount: number;
  billableSeatCount: number;
  seatDiscountCents: number;
  status: 'active' | 'trial' | 'cancelled' | 'expired';
  isGranted?: boolean;
  grantExpiresAt?: string | null;
  trialEndsAt: string | null;
  paddleSubscriptionId: string | null;
}

export interface BillingHistoryMonth {
  month: string;
  totalCents: number;
  lineItems: BillingLineItem[];
  status: 'paid' | 'draft';
}

// ── Helpers ───────────────────────────────────────────────────────────────────

/** Convert cents to a display string: 400 → "$4.00" */
export function centsToDisplay(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

/** Format a date range for line items, e.g. "Apr 1 – 18, 2026" */
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

/** Days remaining in current trial */
export function trialDaysLeft(trialEndsAt: string | null): number {
  if (!trialEndsAt) return 0;
  const diff = new Date(trialEndsAt).getTime() - Date.now();
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)));
}

export interface SeatProration {
  daysRemaining: number;
  daysInMonth: number;
  proratedCents: number;
  nextMonthName: string;
}

/**
 * Compute the prorated seat cost for the remainder of the current billing month.
 * Matches the backend's computeLineItem formula exactly.
 */
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

// ── Query hooks ───────────────────────────────────────────────────────────────

interface MySubscriptionResponse {
  subscription: SubscriptionPayload | null;
}

interface ListPlansResponse {
  plans: PlanPayload[];
}

interface BillingDetailResponse {
  billing: BillingDetailPayload | null;
}

interface BillingHistoryResponse {
  history: BillingHistoryMonth[];
}

interface SetSubscriptionRequest {
  planSlug: string;
}

interface CheckoutRequest {
  planSlug: string;
  email?: string;
}

interface CheckoutResponse {
  checkoutUrl: string;
}

export const useMySubscription = () =>
  useQuery({
    queryKey: ['subscription'],
    queryFn: () => unwrap(safe.get<MySubscriptionResponse>('/subscriptions/my')),
    select: (data) => data.subscription,
  });

export const useListPlans = () =>
  useQuery({
    queryKey: ['subscription', 'plans'],
    queryFn: () => unwrap(safe.get<ListPlansResponse>('/subscriptions/plans')),
    select: (data) => data.plans,
  });

export const useBillingDetail = () =>
  useQuery({
    queryKey: ['subscription', 'billing'],
    queryFn: () => unwrap(safe.get<BillingDetailResponse>('/subscriptions/billing')),
    select: (data) => data.billing,
  });

export const useBillingHistory = () =>
  useQuery({
    queryKey: ['subscription', 'billing', 'history'],
    queryFn: () => unwrap(safe.get<BillingHistoryResponse>('/subscriptions/billing/history')),
    select: (data) => data.history,
  });

export const useSetSubscription = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: SetSubscriptionRequest) =>
      unwrap(safe.post('/subscriptions/my', { plan_slug: data.planSlug })),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['subscription'] });
      toast.success('Subscription updated');
    },
    onError: () => toast.error('Failed to update subscription'),
  });
};

/** Initiates a Paddle checkout and returns the hosted checkout URL. */
export const useCheckout = () =>
  useMutation<CheckoutResponse, Error, CheckoutRequest>({
    mutationFn: (data: CheckoutRequest) =>
      unwrap(safe.post<CheckoutResponse>('/subscriptions/checkout', data)),
  });
