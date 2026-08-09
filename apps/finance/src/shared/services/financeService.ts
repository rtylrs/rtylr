import { safe } from '../lib/api';
import { plReportQueryParams } from './financeApi';

export interface InvoicePayload {
  customerId?: number;
  customerName?: string;
  items?: Array<{ description: string; quantity: number; unitPrice: number }>;
  dueDate?: string;
  notes?: string;
  [key: string]: unknown;
}

export interface InvoiceFilterParams {
  status?: string;
  start?: string;
  end?: string;
  limit?: number;
  offset?: number;
}

export interface PaymentPayload {
  amount: number;
  method?: string;
  paidAt?: string;
  reference?: string;
  [key: string]: unknown;
}

export interface ExpensePayload {
  amount: number;
  category?: string;
  description?: string;
  date?: string;
  supplierId?: number;
  [key: string]: unknown;
}

export interface ExpenseFilterParams {
  category?: string;
  start?: string;
  end?: string;
  limit?: number;
  offset?: number;
}

export interface TaxConfigPayload {
  rate?: number;
  inclusive?: boolean;
  name?: string;
  [key: string]: unknown;
}

export interface PLFilterParams {
  start?: string;
  end?: string;
  groupBy?: string;
}

export const financeService = {
  getDashboard: () => safe.get('/finance/dashboard'),

  getInvoices: (params?: InvoiceFilterParams) => safe.get('/finance/invoices', { params }),
  createInvoice: (data: InvoicePayload) => safe.post('/finance/invoices', data),
  getInvoice: (id: number) => safe.get('/finance/invoices/detail', { params: { id } }),
  updateInvoice: (id: number, data: Partial<InvoicePayload>) =>
    safe.put('/finance/invoices', { id, ...data }),
  deleteInvoice: (id: number) => safe.delete('/finance/invoices', { data: { id } }),
  updateInvoiceStatus: (id: number, status: string) =>
    safe.put('/finance/invoices/status', { id, status }),
  recordPayment: (id: number, data: PaymentPayload) =>
    safe.post('/finance/invoices/payment', { invoiceId: id, ...data }),

  getExpenses: (params?: ExpenseFilterParams) => safe.get('/finance/expenses', { params }),
  createExpense: (data: ExpensePayload) => safe.post('/finance/expenses', data),
  updateExpense: (id: number, data: Partial<ExpensePayload>) =>
    safe.put(`/finance/expenses/${id}`, data),
  deleteExpense: (id: number) => safe.delete(`/finance/expenses/${id}`),
  getExpenseCategories: () => safe.get('/finance/expenses/categories'),

  getTaxConfig: () => safe.get('/finance/tax'),
  updateTaxConfig: (data: TaxConfigPayload) => safe.put('/finance/tax', data),
  getPLReport: (params?: PLFilterParams) =>
    safe.get('/finance/report/v1/pl', { params: plReportQueryParams(params) }),
};
