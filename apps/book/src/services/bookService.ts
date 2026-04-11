import api from '../lib/api';

export const bookService = {
  // Appointments
  getAppointments: (params?: { search?: string; status?: string; date?: string; limit?: number; offset?: number }) =>
    api.get('/book/appointments', { params }),
  getAppointment: (id: number) => api.get(`/book/appointments/${id}`),
  createAppointment: (data: any) => api.post('/book/appointments', data),
  updateAppointment: (id: number, data: any) => api.put(`/book/appointments/${id}`, data),
  cancelAppointment: (id: number) => api.put(`/book/appointments/${id}/cancel`),
  deleteAppointment: (id: number) => api.delete(`/book/appointments/${id}`),

  // Calendar
  getCalendarEvents: (params?: { start?: string; end?: string; staffId?: number }) =>
    api.get('/book/calendar', { params }),

  // Dashboard
  getDashboardStats: () => api.get('/book/dashboard/stats'),
  getTodayAppointments: () => api.get('/book/dashboard/today'),

  // Services
  getServices: (params?: { search?: string; limit?: number; offset?: number }) =>
    api.get('/book/services', { params }),
  getService: (id: number) => api.get(`/book/services/${id}`),
  createService: (data: any) => api.post('/book/services', data),
  updateService: (id: number, data: any) => api.put(`/book/services/${id}`, data),
  deleteService: (id: number) => api.delete(`/book/services/${id}`),

  // Resources
  getResources: (params?: { search?: string; type?: string; limit?: number; offset?: number }) =>
    api.get('/book/resources', { params }),
  getResource: (id: number) => api.get(`/book/resources/${id}`),
  createResource: (data: any) => api.post('/book/resources', data),
  updateResource: (id: number, data: any) => api.put(`/book/resources/${id}`, data),
  deleteResource: (id: number) => api.delete(`/book/resources/${id}`),

  // Waitlist
  getWaitlist: (params?: { search?: string; status?: string; limit?: number; offset?: number }) =>
    api.get('/book/waitlist', { params }),
  addToWaitlist: (data: any) => api.post('/book/waitlist', data),
  notifyWaitlist: (id: number) => api.post(`/book/waitlist/${id}/notify`),
  removeFromWaitlist: (id: number) => api.delete(`/book/waitlist/${id}`),

  // Customers
  getCustomers: (params?: { search?: string; limit?: number; offset?: number }) =>
    api.get('/book/customers', { params }),
  getCustomer: (id: number) => api.get(`/book/customers/${id}`),

  // Settings
  getSettings: () => api.get('/book/settings'),
  updateSettings: (data: any) => api.put('/book/settings', data),
};
