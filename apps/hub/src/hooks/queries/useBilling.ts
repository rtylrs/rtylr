import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';

export interface BillingInfo {
  plan: 'starter' | 'pro' | 'enterprise';
  status: 'active' | 'past_due' | 'canceled' | 'trialing';
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

export interface Invoice {
  id: string;
  number: string;
  amount: number;
  currency: string;
  status: 'paid' | 'open' | 'void' | 'refunded';
  createdAt: string;
  pdfUrl?: string;
}

const MOCK_BILLING: BillingInfo = {
  plan: 'pro',
  status: 'active',
  renewsAt: new Date(Date.now() + 18 * 86400e3).toISOString(),
  amount: 79,
  currency: 'USD',
  paymentMethod: { brand: 'Visa', lastFour: '4242', expMonth: 11, expYear: 2028 },
  usage: {
    seats:     { used: 6,    limit: 15 },
    orders:    { used: 1284, limit: 5000 },
    apiCalls:  { used: 41200, limit: 100000 },
    storageGb: { used: 2.4,  limit: 10 },
  },
};

const MOCK_INVOICES: Invoice[] = [
  { id: 'inv_1', number: 'INV-2026-003', amount: 79, currency: 'USD', status: 'paid', createdAt: new Date(Date.now() - 10 * 86400e3).toISOString() },
  { id: 'inv_2', number: 'INV-2026-002', amount: 79, currency: 'USD', status: 'paid', createdAt: new Date(Date.now() - 40 * 86400e3).toISOString() },
  { id: 'inv_3', number: 'INV-2026-001', amount: 79, currency: 'USD', status: 'paid', createdAt: new Date(Date.now() - 70 * 86400e3).toISOString() },
];

export const useBilling = () => {
  return useQuery<BillingInfo>({
    queryKey: ['billing'],
    queryFn: async () => {
      try {
        const res = await api.get('/billing/summary');
        return res.data as BillingInfo;
      } catch {
        return MOCK_BILLING;
      }
    },
  });
};

export const useInvoices = () => {
  return useQuery<Invoice[]>({
    queryKey: ['invoices'],
    queryFn: async () => {
      try {
        const res = await api.get('/billing/invoices');
        return (res.data?.invoices ?? res.data) as Invoice[];
      } catch {
        return MOCK_INVOICES;
      }
    },
  });
};
