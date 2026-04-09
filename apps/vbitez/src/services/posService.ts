import api from '../lib/api';

export const posService = {
  // Tables
  getTables: () => api.get('/pos/tables'),
  createTable: (data: any) => api.post('/pos/tables', data),
  updateTable: (id: number, data: any) => api.put(`/pos/tables/${id}`, data),
  deleteTable: (id: number) => api.delete(`/pos/tables/${id}`),

  // Orders
  getOrders: (params?: any) => api.get('/pos/orders', { params }),
  getOrder: (id: string | number) => api.get(`/pos/orders/${id}`),
  createOrder: (data: any) => api.post('/pos/orders', data),
  updateOrder: (id: number | string, data: any) => api.put(`/pos/orders/${id}`, data),
  chargeOrder: (id: number | string, data: any) => api.post(`/pos/orders/${id}/charge`, data),
  voidOrder: (id: number | string, data: any) => api.post(`/pos/orders/${id}/void`, data),
  refundOrder: (id: number | string, data: any) => api.post(`/pos/orders/${id}/refund`, data),
  applyDiscount: (id: number | string, data: any) =>
    api.post(`/pos/orders/${id}/discount`, data),
  splitOrder: (id: number | string, data: any) => api.post(`/pos/orders/${id}/split`, data),
  holdOrder: (id: number | string) => api.post(`/pos/orders/${id}/hold`),
  releaseOrder: (id: number | string) => api.post(`/pos/orders/${id}/release`),

  // Menu
  getMenuItems: (params?: any) => api.get('/pos/menu/items', { params }),
  getModifiers: () => api.get('/pos/menu/modifiers'),
  getCombos: () => api.get('/pos/menu/combos'),
  updateItem: (id: string | number, data: any) => api.put(`/pos/menu/items/${id}`, data),

  // Sessions
  getSessions: () => api.get('/pos/sessions'),
  openSession: (data: any) => api.post('/pos/sessions/open', data),
  closeSession: (id: number | string, data: any) => api.post(`/pos/sessions/${id}/close`, data),

  // Locations
  getLocations: () => api.get('/pos/locations'),
  createLocation: (data: any) => api.post('/pos/locations', data),

  // Gift cards
  getGiftCards: () => api.get('/pos/gift-cards'),
  createGiftCard: (data: any) => api.post('/pos/gift-cards', data),
  redeemGiftCard: (code: string, data: any) => api.post(`/pos/gift-cards/${code}/redeem`, data),
  getGiftCardBalance: (code: string) => api.get(`/pos/gift-cards/${code}`),

  // KDS
  getKDSOrders: (station?: string) => api.get('/pos/kds', { params: { station } }),
  updateKDSStatus: (id: number | string, data: any) => api.put(`/pos/kds/${id}`, data),

  // Customers
  searchCustomers: (q: string) => api.get('/pos/customers', { params: { q } }),
  createCustomer: (data: any) => api.post('/pos/customers', data),
  getCustomer: (id: string | number) => api.get(`/pos/customers/${id}`),

  // Receipts
  getReceipts: (params?: any) => api.get('/pos/receipts', { params }),
  emailReceipt: (id: string | number, email: string) =>
    api.post(`/pos/receipts/${id}/email`, { email }),

  // Day summary
  getDaySummary: (date?: string) => api.get('/pos/day-summary', { params: { date } }),

  // Manager override
  verifyManagerPin: (pin: string) => api.post('/pos/manager/verify', { pin }),
};
