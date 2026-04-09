import api from '../lib/api';

export const financeService = {
  getDashboard: () => api.get('/finance/dashboard'),

  getInvoices: (params?: any) => api.get('/finance/invoices', { params }),
  createInvoice: (data: any) => api.post('/finance/invoices', data),
  getInvoice: (id: number) => api.get(`/finance/invoices/${id}`),
  updateInvoice: (id: number, data: any) => api.put(`/finance/invoices/${id}`, data),
  deleteInvoice: (id: number) => api.delete(`/finance/invoices/${id}`),
  updateInvoiceStatus: (id: number, status: string) => api.put(`/finance/invoices/${id}/status`, { status }),
  recordPayment: (id: number, data: any) => api.post(`/finance/invoices/${id}/payment`, data),

  getExpenses: (params?: any) => api.get('/finance/expenses', { params }),
  createExpense: (data: any) => api.post('/finance/expenses', data),
  updateExpense: (id: number, data: any) => api.put(`/finance/expenses/${id}`, data),
  deleteExpense: (id: number) => api.delete(`/finance/expenses/${id}`),
  getExpenseCategories: () => api.get('/finance/expenses/categories'),

  getTaxConfig: () => api.get('/finance/tax'),
  updateTaxConfig: (data: any) => api.put('/finance/tax', data),
  getPLReport: (params?: any) => api.get('/finance/pl', { params }),
};
