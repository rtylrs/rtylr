import api from '../lib/api';

export const posService = {
  // Tables
  getTables: () => api.get('/pos/tables'),
  createTable: (data: any) => api.post('/pos/tables', data),
  updateTable: (id: number, data: any) => api.put(`/pos/tables/${id}`, data),
  deleteTable: (id: number) => api.delete(`/pos/tables/${id}`),

  // Orders
  getOrders: (params?: any) => api.get('/pos/orders', { params }),
  createOrder: (data: any) => api.post('/pos/orders', data),
  updateOrder: (id: number, data: any) => api.put(`/pos/orders/${id}`, data),
  chargeOrder: (id: number, data: any) => api.post(`/pos/orders/${id}/charge`, data),

  // Sessions
  getSessions: () => api.get('/pos/sessions'),
  openSession: (data: any) => api.post('/pos/sessions/open', data),
  closeSession: (id: number, data: any) => api.post(`/pos/sessions/${id}/close`, data),

  // Locations
  getLocations: () => api.get('/pos/locations'),
  createLocation: (data: any) => api.post('/pos/locations', data),
};
