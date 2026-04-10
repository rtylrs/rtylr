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

export const useBilling = () => {
  return useQuery<BillingInfo>({
    queryKey: ['billing'],
    queryFn: async () => {
      const res = await api.get('/billing/summary');
      return res.data as BillingInfo;
    },
  });
};

export const useInvoices = () => {
  return useQuery<Invoice[]>({
    queryKey: ['invoices'],
    queryFn: async () => {
      const res = await api.get('/billing/invoices');
      return (res.data?.invoices ?? res.data) as Invoice[];
    },
  });
};
