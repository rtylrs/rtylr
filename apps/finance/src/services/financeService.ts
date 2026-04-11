import { safe } from '../lib/api';

export const financeService = {
  getDashboard: () => safe.get('/finance/dashboard'),

  getInvoices: (params?: any) => safe.get('/finance/invoices', { params }),
  createInvoice: (data: any) => safe.post('/finance/invoices', data),
  getInvoice: (id: number) => safe.get(`/finance/invoices/${id}`),
  updateInvoice: (id: number, data: any) => safe.put(`/finance/invoices/${id}`, data),
  deleteInvoice: (id: number) => safe.delete(`/finance/invoices/${id}`),
  updateInvoiceStatus: (id: number, status: string) => safe.put(`/finance/invoices/${id}/status`, { status }),
  recordPayment: (id: number, data: any) => safe.post(`/finance/invoices/${id}/payment`, data),

  getExpenses: (params?: any) => safe.get('/finance/expenses', { params }),
  createExpense: (data: any) => safe.post('/finance/expenses', data),
  updateExpense: (id: number, data: any) => safe.put(`/finance/expenses/${id}`, data),
  deleteExpense: (id: number) => safe.delete(`/finance/expenses/${id}`),
  getExpenseCategories: () => safe.get('/finance/expenses/categories'),

  getTaxConfig: () => safe.get('/finance/tax'),
  updateTaxConfig: (data: any) => safe.put('/finance/tax', data),
  getPLReport: (params?: any) => safe.get('/finance/pl', { params }),
};
