import api from '../lib/api';

export const crmService = {
  getDashboard: () => api.get('/crm/dashboard'),

  getCustomers: (params?: any) => api.get('/crm/customers', { params }),
  createCustomer: (data: any) => api.post('/crm/customers', data),
  getCustomer: (id: number) => api.get(`/crm/customers/${id}`),
  updateCustomer: (id: number, data: any) => api.put(`/crm/customers/${id}`, data),
  deleteCustomer: (id: number) => api.delete(`/crm/customers/${id}`),
  recordVisit: (data: any) => api.post('/crm/customers/visit', data),

  getLoyaltyTransactions: (params?: any) => api.get('/crm/loyalty/transactions', { params }),
  awardPoints: (data: any) => api.post('/crm/loyalty/award', data),
  redeemPoints: (data: any) => api.post('/crm/loyalty/redeem', data),
  getLoyaltyConfig: () => api.get('/crm/loyalty/config'),
  updateLoyaltyConfig: (data: any) => api.put('/crm/loyalty/config', data),

  getSegments: () => api.get('/crm/segments'),
  createSegment: (data: any) => api.post('/crm/segments', data),
  updateSegment: (id: number, data: any) => api.put(`/crm/segments/${id}`, data),
  deleteSegment: (id: number) => api.delete(`/crm/segments/${id}`),

  getCampaigns: () => api.get('/crm/campaigns'),
  createCampaign: (data: any) => api.post('/crm/campaigns', data),
  updateCampaign: (id: number, data: any) => api.put(`/crm/campaigns/${id}`, data),
  updateCampaignStatus: (id: number, status: string) => api.put(`/crm/campaigns/${id}/status`, { status }),
};
