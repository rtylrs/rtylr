import { safe } from '../lib/api';

export const bookService = {
  // Appointments
  getAppointments: (params?: {
    search?: string;
    status?: string;
    date?: string;
    limit?: number;
    offset?: number;
  }) => safe.get('/book/appointments', { params }),
  getAppointment: (id: number) => safe.get(`/book/appointments/${id}`),
  createAppointment: (data: any) => safe.post('/book/appointments', data),
  updateAppointment: (id: number, data: any) => safe.put(`/book/appointments/${id}`, data),
  cancelAppointment: (id: number) => safe.put(`/book/appointments/${id}/cancel`),
  deleteAppointment: (id: number) => safe.delete(`/book/appointments/${id}`),

  // Calendar
  getCalendarEvents: (params?: { start?: string; end?: string; staffId?: number }) =>
    safe.get('/book/calendar', { params }),

  // Dashboard
  getDashboardStats: () => safe.get('/book/dashboard/stats'),
  getTodayAppointments: () => safe.get('/book/dashboard/today'),

  // Services
  getServices: (params?: { search?: string; limit?: number; offset?: number }) =>
    safe.get('/book/services', { params }),
  getService: (id: number) => safe.get(`/book/services/${id}`),
  createService: (data: any) => safe.post('/book/services', data),
  updateService: (id: number, data: any) => safe.put(`/book/services/${id}`, data),
  deleteService: (id: number) => safe.delete(`/book/services/${id}`),

  // Resources
  getResources: (params?: { search?: string; type?: string; limit?: number; offset?: number }) =>
    safe.get('/book/resources', { params }),
  getResource: (id: number) => safe.get(`/book/resources/${id}`),
  createResource: (data: any) => safe.post('/book/resources', data),
  updateResource: (id: number, data: any) => safe.put(`/book/resources/${id}`, data),
  deleteResource: (id: number) => safe.delete(`/book/resources/${id}`),

  // Waitlist
  getWaitlist: (params?: { search?: string; status?: string; limit?: number; offset?: number }) =>
    safe.get('/book/waitlist', { params }),
  addToWaitlist: (data: any) => safe.post('/book/waitlist', data),
  notifyWaitlist: (id: number) => safe.post(`/book/waitlist/${id}/notify`),
  removeFromWaitlist: (id: number) => safe.delete(`/book/waitlist/${id}`),

  // Customers
  getCustomers: (params?: { search?: string; limit?: number; offset?: number }) =>
    safe.get('/book/customers', { params }),
  getCustomer: (id: number) => safe.get(`/book/customers/${id}`),

  // Settings
  getSettings: () => safe.get('/book/settings'),
  updateSettings: (data: any) => safe.put('/book/settings', data),
};
