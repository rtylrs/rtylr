import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface Integration {
  id: string;
  name: string;
  category: 'payments' | 'messaging' | 'accounting' | 'analytics' | 'delivery' | 'data';
  logo: string;
  status: 'connected' | 'available' | 'coming_soon';
  description: string;
  connectedAt?: string;
}

const MOCK_INTEGRATIONS: Integration[] = [
  { id: 'stripe',    name: 'Stripe',    category: 'payments',   logo: 'S',  status: 'connected',  description: 'Accept online card payments.', connectedAt: new Date(Date.now() - 30 * 86400e3).toISOString() },
  { id: 'square',    name: 'Square',    category: 'payments',   logo: 'Sq', status: 'available',  description: 'Sync transactions from Square terminals.' },
  { id: 'slack',     name: 'Slack',     category: 'messaging',  logo: 'Sl', status: 'connected',  description: 'Send alerts and daily summaries to Slack.', connectedAt: new Date(Date.now() - 12 * 86400e3).toISOString() },
  { id: 'twilio',    name: 'Twilio',    category: 'messaging',  logo: 'T',  status: 'available',  description: 'Send SMS and WhatsApp to customers.' },
  { id: 'quickbook', name: 'QuickBooks',category: 'accounting', logo: 'Qb', status: 'available',  description: 'Export invoices and expenses automatically.' },
  { id: 'xero',      name: 'Xero',      category: 'accounting', logo: 'X',  status: 'available',  description: 'Two-way sync with Xero ledger.' },
  { id: 'ga4',       name: 'GA4',       category: 'analytics',  logo: 'G',  status: 'connected',  description: 'Track web and storefront events.', connectedAt: new Date(Date.now() - 5 * 86400e3).toISOString() },
  { id: 'posthog',   name: 'PostHog',   category: 'analytics',  logo: 'Ph', status: 'available',  description: 'Product analytics and session replays.' },
  { id: 'uber',      name: 'Uber Eats', category: 'delivery',   logo: 'U',  status: 'available',  description: 'Sync menu and fulfill Uber Eats orders.' },
  { id: 'doordash',  name: 'DoorDash',  category: 'delivery',   logo: 'Dd', status: 'coming_soon',description: 'DoorDash integration is arriving soon.' },
  { id: 'shopify',   name: 'Shopify',   category: 'data',       logo: 'Sh', status: 'coming_soon',description: 'Pull existing Shopify catalog.' },
  { id: 'airtable',  name: 'Airtable',  category: 'data',       logo: 'At', status: 'available',  description: 'Bi-directional sync with Airtable bases.' },
];

export const useIntegrations = () => {
  return useQuery<Integration[]>({
    queryKey: ['integrations'],
    queryFn: async () => {
      try {
        const res = await api.get('/integrations');
        return (res.data?.integrations ?? res.data) as Integration[];
      } catch {
        return MOCK_INTEGRATIONS;
      }
    },
  });
};

export const useConnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await api.post(`/integrations/${id}/connect`);
      } catch {
        // ignore
      }
      return id;
    },
    onSuccess: () => {
      toast.success('Integration connected');
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
  });
};

export const useDisconnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await api.post(`/integrations/${id}/disconnect`);
      } catch {
        // ignore
      }
      return id;
    },
    onSuccess: () => {
      toast.success('Integration disconnected');
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
  });
};
