import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { subscriptionClient } from '../../lib/clients';

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

type ApiScalar = string | number | boolean | null | undefined;

type RawSubscriptionPayload = Partial<SubscriptionPayload> & {
  plan_name?: string;
  plan_slug?: string;
  price_per_seat_cents?: ApiScalar;
  is_granted?: ApiScalar;
  seat_count?: ApiScalar;
  included_seat_limit?: ApiScalar;
  included_seat_count?: ApiScalar;
  billable_seat_count?: ApiScalar;
  grant_expires_at?: string | null;
  trial_ends_at?: string | null;
  next_billing_date?: string | null;
  paddle_subscription_id?: string | null;
};

type RawPlanPayload = Partial<PlanPayload> & {
  price_per_seat_cents?: ApiScalar;
  annual_price_per_seat_cents?: ApiScalar;
  trial_days?: ApiScalar;
  max_seats?: ApiScalar;
  max_locations?: ApiScalar;
  ai_credits_per_month?: ApiScalar;
  tier_order?: ApiScalar;
  is_solo?: ApiScalar;
  is_featured?: ApiScalar;
  is_active?: ApiScalar;
};

type RawBillingLineItem = Partial<BillingLineItem> & {
  period_start?: string;
  period_end?: string | null;
  seat_count?: ApiScalar;
  included_seat_limit?: ApiScalar;
  included_seat_count?: ApiScalar;
  billable_seat_count?: ApiScalar;
  included_cents?: ApiScalar;
  price_per_seat_cents?: ApiScalar;
  amount_cents?: ApiScalar;
  is_granted?: ApiScalar;
  grant_expires_at?: string | null;
  addon_id?: string;
  unit_price_cents?: ApiScalar;
};

type RawBillingDetailPayload = Partial<BillingDetailPayload> & {
  billing_month?: string;
  next_billing_date?: string | null;
  line_items?: RawBillingLineItem[];
  total_cents?: ApiScalar;
  plan_name?: string;
  price_per_seat_cents?: ApiScalar;
  current_seat_count?: ApiScalar;
  included_seat_limit?: ApiScalar;
  included_seat_count?: ApiScalar;
  billable_seat_count?: ApiScalar;
  seat_discount_cents?: ApiScalar;
  is_granted?: ApiScalar;
  grant_expires_at?: string | null;
  trial_ends_at?: string | null;
  paddle_subscription_id?: string | null;
};

type RawBillingHistoryMonth = Partial<BillingHistoryMonth> & {
  billing_month?: string;
  billingMonth?: string;
  month_iso?: string;
  period?: string;
  date?: string;
  total_cents?: ApiScalar;
  line_items?: RawBillingLineItem[];
};

// ── Helpers ───────────────────────────────────────────────────────────────────

/** Convert cents to a display string: 400 → "$4.00" */
export function centsToDisplay(cents: number): string {
  const safeCents = Number.isFinite(cents) ? cents : 0;
  return `$${(safeCents / 100).toFixed(2)}`;
}

function num(value: ApiScalar, fallback = 0): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function nullableNum(value: ApiScalar): number | null {
  if (value === null || value === undefined || value === '') return null;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function bool(value: ApiScalar, fallback = false): boolean {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase();
    if (['true', '1', 'yes'].includes(normalized)) return true;
    if (['false', '0', 'no', ''].includes(normalized)) return false;
  }
  return fallback;
}

function stringList(value: unknown): string[] {
  return Array.isArray(value) ? value.map((item) => String(item)) : [];
}

function normalizeSubscription(
  raw: RawSubscriptionPayload | null | undefined
): SubscriptionPayload | null {
  if (!raw) return null;
  return {
    id: num(raw.id),
    planName: raw.planName ?? raw.plan_name ?? '',
    planSlug: raw.planSlug ?? raw.plan_slug ?? '',
    apps: stringList(raw.apps),
    pricePerSeatCents: num(raw.pricePerSeatCents ?? raw.price_per_seat_cents),
    status: raw.status ?? 'active',
    isGranted: bool(raw.isGranted ?? raw.is_granted),
    seatCount: num(raw.seatCount ?? raw.seat_count),
    includedSeatLimit: num(raw.includedSeatLimit ?? raw.included_seat_limit),
    includedSeatCount: num(raw.includedSeatCount ?? raw.included_seat_count),
    billableSeatCount: num(raw.billableSeatCount ?? raw.billable_seat_count),
    grantExpiresAt: raw.grantExpiresAt ?? raw.grant_expires_at ?? null,
    trialEndsAt: raw.trialEndsAt ?? raw.trial_ends_at ?? null,
    nextBillingDate: raw.nextBillingDate ?? raw.next_billing_date ?? null,
    paddleSubscriptionId: raw.paddleSubscriptionId ?? raw.paddle_subscription_id ?? null,
  };
}

function normalizePlan(raw: RawPlanPayload): PlanPayload {
  return {
    id: num(raw.id),
    name: raw.name ?? '',
    slug: raw.slug ?? '',
    apps: stringList(raw.apps),
    pricePerSeatCents: num(raw.pricePerSeatCents ?? raw.price_per_seat_cents),
    annualPricePerSeatCents: num(raw.annualPricePerSeatCents ?? raw.annual_price_per_seat_cents),
    trialDays: num(raw.trialDays ?? raw.trial_days),
    tagline: raw.tagline ?? '',
    maxSeats: nullableNum(raw.maxSeats ?? raw.max_seats),
    maxLocations: nullableNum(raw.maxLocations ?? raw.max_locations),
    aiCreditsPerMonth: num(raw.aiCreditsPerMonth ?? raw.ai_credits_per_month),
    features: stringList(raw.features),
    tierOrder: num(raw.tierOrder ?? raw.tier_order),
    isSolo: bool(raw.isSolo ?? raw.is_solo),
    isFeatured: bool(raw.isFeatured ?? raw.is_featured),
    isActive: bool(raw.isActive ?? raw.is_active, true),
  };
}

function normalizeLineItem(raw: RawBillingLineItem): BillingLineItem {
  return {
    periodStart: raw.periodStart ?? raw.period_start ?? '',
    periodEnd: raw.periodEnd ?? raw.period_end ?? null,
    seatCount: num(raw.seatCount ?? raw.seat_count),
    includedSeatLimit: num(raw.includedSeatLimit ?? raw.included_seat_limit),
    includedSeatCount: num(raw.includedSeatCount ?? raw.included_seat_count),
    billableSeatCount: num(raw.billableSeatCount ?? raw.billable_seat_count),
    includedCents: num(raw.includedCents ?? raw.included_cents),
    pricePerSeatCents: num(raw.pricePerSeatCents ?? raw.price_per_seat_cents),
    amountCents: num(raw.amountCents ?? raw.amount_cents),
    type: raw.type,
    description: raw.description,
    isGranted: bool(raw.isGranted ?? raw.is_granted),
    grantExpiresAt: raw.grantExpiresAt ?? raw.grant_expires_at ?? null,
    addonId: raw.addonId ?? raw.addon_id,
    quantity: num(raw.quantity),
    unitPriceCents: num(raw.unitPriceCents ?? raw.unit_price_cents),
  };
}

function normalizeBilling(
  raw: RawBillingDetailPayload | null | undefined
): BillingDetailPayload | null {
  if (!raw) return null;
  const lineItems = raw.lineItems ?? raw.line_items ?? [];
  return {
    billingMonth: raw.billingMonth ?? raw.billing_month ?? '',
    nextBillingDate: raw.nextBillingDate ?? raw.next_billing_date ?? null,
    lineItems: lineItems.map(normalizeLineItem),
    totalCents: num(raw.totalCents ?? raw.total_cents),
    planName: raw.planName ?? raw.plan_name ?? '',
    pricePerSeatCents: num(raw.pricePerSeatCents ?? raw.price_per_seat_cents),
    currentSeatCount: num(raw.currentSeatCount ?? raw.current_seat_count),
    includedSeatLimit: num(raw.includedSeatLimit ?? raw.included_seat_limit),
    includedSeatCount: num(raw.includedSeatCount ?? raw.included_seat_count),
    billableSeatCount: num(raw.billableSeatCount ?? raw.billable_seat_count),
    seatDiscountCents: num(raw.seatDiscountCents ?? raw.seat_discount_cents),
    status: raw.status ?? 'active',
    isGranted: bool(raw.isGranted ?? raw.is_granted),
    grantExpiresAt: raw.grantExpiresAt ?? raw.grant_expires_at ?? null,
    trialEndsAt: raw.trialEndsAt ?? raw.trial_ends_at ?? null,
    paddleSubscriptionId: raw.paddleSubscriptionId ?? raw.paddle_subscription_id ?? null,
  };
}

function normalizeBillingHistory(raw: RawBillingHistoryMonth): BillingHistoryMonth {
  const lineItems = raw.lineItems ?? raw.line_items ?? [];
  return {
    month:
      raw.month ??
      raw.billingMonth ??
      raw.billing_month ??
      raw.month_iso ??
      raw.period ??
      raw.date ??
      '',
    totalCents: num(raw.totalCents ?? raw.total_cents),
    lineItems: lineItems.map(normalizeLineItem),
    status: raw.status ?? 'draft',
  };
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
  subscription: RawSubscriptionPayload | null;
}

interface ListPlansResponse {
  plans: RawPlanPayload[];
}

interface BillingDetailResponse {
  billing: RawBillingDetailPayload | null;
}


interface SetSubscriptionRequest {
  planSlug: string;
}

interface CheckoutRequest {
  planSlug: string;
  billingCycle?: 'monthly' | 'yearly';
  seatCount?: number;
  addons?: string[];
  email?: string;
}

interface CheckoutResponse {
  checkoutUrl: string;
}

export const useMySubscription = () =>
  useQuery({
    queryKey: ['subscription'],
    queryFn: () => subscriptionClient.getMySubscription({}) as Promise<MySubscriptionResponse>,
    select: (data) => normalizeSubscription(data.subscription),
  });

export const useListPlans = () =>
  useQuery({
    queryKey: ['subscription', 'plans'],
    queryFn: () => subscriptionClient.listPlans({}) as Promise<ListPlansResponse>,
    select: (data) => (data.plans ?? []).map(normalizePlan),
  });

export const useBillingDetail = () =>
  useQuery({
    queryKey: ['subscription', 'billing'],
    queryFn: () => subscriptionClient.getBillingDetail({}) as Promise<BillingDetailResponse>,
    select: (data) => normalizeBilling(data.billing),
  });

export const useBillingHistory = () =>
  useQuery({
    queryKey: ['subscription', 'billing', 'history'],
    queryFn: () => subscriptionClient.getBillingHistory({}),
    select: (data) =>
      (data.history ?? []).map((month) =>
        normalizeBillingHistory(month as unknown as RawBillingHistoryMonth)
      ),
  });

import i18next from 'i18next';

export const useSetSubscription = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: SetSubscriptionRequest) =>
      subscriptionClient.upsertSubscription({ planSlug: data.planSlug }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['subscription'] });
      toast.success(i18next.t('common.toasts.subscriptionUpdated', 'Subscription updated'));
    },
    onError: () =>
      toast.error(
        i18next.t('common.toasts.failedUpdateSubscription', 'Failed to update subscription')
      ),
  });
};

/** Initiates a Paddle checkout and returns the hosted checkout URL. */
export const useCheckout = () =>
  useMutation<CheckoutResponse, Error, CheckoutRequest>({
    mutationFn: (data: CheckoutRequest) =>
      subscriptionClient.createCheckout({
        planSlug: data.planSlug,
        email: data.email ?? '',
        billingCycle: data.billingCycle,
        seatCount: data.seatCount,
        addons: data.addons ?? [],
      }) as Promise<CheckoutResponse>,
  });
