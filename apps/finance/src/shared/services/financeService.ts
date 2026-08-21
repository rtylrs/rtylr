import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import { financeApi, plReportQueryParams } from './financeApi';

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
  getDashboard: () => financeApi.getDashboard(),

  getInvoices: (params?: InvoiceFilterParams) => financeApi.getInvoices(params),
  createInvoice: (data: InvoicePayload) =>
    financeApi.createInvoice(data as unknown as CreateInvoiceRequest),
  getInvoice: (id: number) => financeApi.getInvoiceDetail(String(id)),
  updateInvoice: (id: number, data: Partial<InvoicePayload>) =>
    financeApi.updateInvoice(String(id), data),
  deleteInvoice: (id: number) => financeApi.deleteInvoice(String(id)),
  updateInvoiceStatus: (id: number, status: string) =>
    financeApi.updateInvoiceStatus(String(id), status as never),
  recordPayment: (id: number, data: PaymentPayload) =>
    financeApi.recordPayment(String(id), data.amount, data),

  getExpenses: (params?: ExpenseFilterParams) => financeApi.getExpenses(params),
  createExpense: (data: ExpensePayload) => financeApi.createExpense(data),
  updateExpense: (id: number, data: Partial<ExpensePayload>) =>
    financeApi.updateExpense(String(id), data),
  deleteExpense: (id: number) => financeApi.deleteExpense(String(id)),
  getExpenseCategories: () => financeApi.getExpenseCategories(),

  getTaxConfig: () => financeApi.getTaxConfig(),
  updateTaxConfig: (data: TaxConfigPayload) => financeApi.updateTaxConfig(data),
  getPLReport: (params?: PLFilterParams) => financeApi.getPL(plReportQueryParams(params).period),
};
