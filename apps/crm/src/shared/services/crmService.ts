import { safe } from '../lib/api';

export interface CustomerPayload {
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  notes?: string;
  [key: string]: unknown;
}

export interface CustomerFilterParams {
  search?: string;
  segment?: string;
  limit?: number;
  offset?: number;
}

export interface LoyaltyFilterParams {
  customerId?: number;
  start?: string;
  end?: string;
  limit?: number;
  offset?: number;
}

export interface LoyaltyPayload {
  customerId: number;
  points: number;
  reason?: string;
  [key: string]: unknown;
}

export interface LoyaltyConfigPayload {
  pointsPerDollar?: number;
  redeemRate?: number;
  expiryDays?: number;
  [key: string]: unknown;
}

export interface SegmentPayload {
  name: string;
  conditions?: Record<string, unknown>[];
  [key: string]: unknown;
}

export interface CampaignPayload {
  name: string;
  type?: string;
  audience?: string;
  message?: string;
  scheduledAt?: string;
  [key: string]: unknown;
}

export interface VisitPayload {
  customerId: number;
  locationId?: number;
  [key: string]: unknown;
}

export const crmService = {
  getDashboard: () => safe.get('/crm/dashboard'),

  getCustomers: (params?: CustomerFilterParams) => safe.get('/crm/customers', { params }),
  createCustomer: (data: CustomerPayload) => safe.post('/crm/customers', data),
  getCustomer: (id: number) => safe.get('/crm/customers/detail', { params: { id } }),
  updateCustomer: (id: number, data: Partial<CustomerPayload>) =>
    safe.put('/crm/customers', { id, ...data }),
  deleteCustomer: (id: number) => safe.delete('/crm/customers', { data: { id } }),
  recordVisit: (data: VisitPayload) => safe.post('/crm/customers/record-visit', data),

  getLoyaltyTransactions: (params?: LoyaltyFilterParams) =>
    safe.get('/crm/loyalty/transactions', { params }),
  awardPoints: (data: LoyaltyPayload) => safe.post('/crm/loyalty/award', data),
  redeemPoints: (data: LoyaltyPayload) => safe.post('/crm/loyalty/redeem', data),
  getLoyaltyConfig: () => safe.get('/crm/loyalty/config'),
  updateLoyaltyConfig: (data: LoyaltyConfigPayload) => safe.put('/crm/loyalty/config', data),

  getSegments: () => safe.get('/crm/segments'),
  createSegment: (data: SegmentPayload) => safe.post('/crm/segments', data),
  updateSegment: (id: number, data: Partial<SegmentPayload>) =>
    safe.put(`/crm/segments/${id}`, data),
  deleteSegment: (id: number) => safe.delete(`/crm/segments/${id}`),

  getCampaigns: () => safe.get('/crm/campaigns'),
  createCampaign: (data: CampaignPayload) => safe.post('/crm/campaigns', data),
  updateCampaign: (id: number, data: Partial<CampaignPayload>) =>
    safe.put(`/crm/campaigns/${id}`, data),
  updateCampaignStatus: (id: number, status: string) =>
    safe.put(`/crm/campaigns/${id}/status`, { status }),
};
