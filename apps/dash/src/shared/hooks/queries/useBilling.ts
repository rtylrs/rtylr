import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { billingSummarySchema, invoiceSchema } from '../../schemas/billing';

export type { Invoice } from '../../schemas/billing';

export interface BillingInfo {
  plan: string;
  status: string;
  renewsAt: string;
  amount: number;
  currency: string;
  paymentMethod?: { brand: string; lastFour: string; expMonth: number; expYear: number };
  usage: {
    seats: { used: number; limit: number };
    orders: { used: number; limit: number };
    apiCalls: { used: number; limit: number };
    storageGb: { used: number; limit: number };
  };
}

const PLAN_LIMITS: Record<
  string,
  { seats: number; orders: number; apiCalls: number; storageGb: number }
> = {
  free: { seats: 5, orders: 500, apiCalls: 10000, storageGb: 1 },
  pro: { seats: 20, orders: 5000, apiCalls: 100000, storageGb: 10 },
  scale: { seats: 999, orders: 999999, apiCalls: 999999, storageGb: 100 },
};

const PLAN_PRICES: Record<string, number> = {
  free: 0,
  pro: 4.99,
  scale: 2.99,
};

export const useBilling = () => {
  return useQuery<BillingInfo>({
    queryKey: ['billing'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/api/hub/billing/v1/summary'));
      const raw = billingSummarySchema.parse(
        (res as { summary?: unknown })?.summary ?? res ?? {}
      );
      const plan = raw.plan || 'free';
      const limits = PLAN_LIMITS[plan] || PLAN_LIMITS.free;

      return {
        plan,
        status: raw.status || 'active',
        renewsAt: raw.nextBillingDate || '',
        amount: PLAN_PRICES[plan] || 0,
        currency: 'USD',
        usage: {
          seats: { used: raw.currentUsage || 1, limit: limits.seats },
          orders: { used: 0, limit: limits.orders },
          apiCalls: { used: 0, limit: limits.apiCalls },
          storageGb: { used: 0, limit: limits.storageGb },
        },
      };
    },
  });
};

export const useInvoices = () => {
  return useQuery({
    queryKey: ['invoices'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/api/hub/billing/v1/invoices'));
      const raw = res?.invoices ?? res ?? [];
      return z.array(invoiceSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};
