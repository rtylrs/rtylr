import { safe } from '../lib/api';

export const crmService = {
  getDashboard: () => safe.get('/crm/dashboard'),

  getCustomers: (params?: any) => safe.get('/crm/customers', { params }),
  createCustomer: (data: any) => safe.post('/crm/customers', data),
  getCustomer: (id: number) => safe.get(`/crm/customers/${id}`),
  updateCustomer: (id: number, data: any) => safe.put(`/crm/customers/${id}`, data),
  deleteCustomer: (id: number) => safe.delete(`/crm/customers/${id}`),
  recordVisit: (data: any) => safe.post('/crm/customers/visit', data),

  getLoyaltyTransactions: (params?: any) => safe.get('/crm/loyalty/transactions', { params }),
  awardPoints: (data: any) => safe.post('/crm/loyalty/award', data),
  redeemPoints: (data: any) => safe.post('/crm/loyalty/redeem', data),
  getLoyaltyConfig: () => safe.get('/crm/loyalty/config'),
  updateLoyaltyConfig: (data: any) => safe.put('/crm/loyalty/config', data),

  getSegments: () => safe.get('/crm/segments'),
  createSegment: (data: any) => safe.post('/crm/segments', data),
  updateSegment: (id: number, data: any) => safe.put(`/crm/segments/${id}`, data),
  deleteSegment: (id: number) => safe.delete(`/crm/segments/${id}`),

  getCampaigns: () => safe.get('/crm/campaigns'),
  createCampaign: (data: any) => safe.post('/crm/campaigns', data),
  updateCampaign: (id: number, data: any) => safe.put(`/crm/campaigns/${id}`, data),
  updateCampaignStatus: (id: number, status: string) =>
    safe.put(`/crm/campaigns/${id}/status`, { status }),
};
