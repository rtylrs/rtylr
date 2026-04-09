import api from '../lib/api';

// ── Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  totalRevenueMTD: number;
  totalExpensesMTD: number;
  netProfit: number;
  outstandingInvoicesTotal: number;
  outstandingInvoicesCount: number;
  totalRevenueMTDChange: string;
  totalExpensesMTDChange: string;
  netProfitChange: string;
  recentTransactions: Transaction[];
  overdueInvoices: OverdueInvoice[];
  barData: BarDataPoint[];
}

export interface Transaction {
  id: string;
  description: string;
  amount: string;
  type: 'credit' | 'debit';
  date: string;
}

export interface OverdueInvoice {
  id: string;
  client: string;
  amount: string;
  daysOverdue: number;
}

export interface BarDataPoint {
  month: string;
  revenue: number;
  expenses: number;
}

export interface Invoice {
  id: string;
  client: string;
  amount: number;
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue';
  dueDate: string;
  issuedDate: string;
}

export interface Expense {
  id: string;
  date: string;
  description: string;
  category: string;
  amount: number;
  receipt: boolean;
}

export interface ExpenseCategory {
  id: number;
  name: string;
  color: string;
}

export interface PLReport {
  period: string;
  revenue: number;
  cogs: number;
  operatingExpenses: number;
  grossProfit: number;
  netProfit: number;
}

export interface TaxConfig {
  id?: number;
  name: string;
  rate: number;
}

export interface TaxPeriod {
  id: string;
  period: string;
  revenue: number;
  taxableAmount: number;
  taxRate: number;
  collected: number;
  status: 'Filed' | 'Pending' | 'Due';
  dueDate: string;
}

export interface TaxReport {
  vatSummary: { collected: number; owed: number; paid: number };
  periods: TaxPeriod[];
  configs: TaxConfig[];
}

// ── Service ──────────────────────────────────────────────────────

export const financeApi = {
  // Dashboard
  getDashboard: () =>
    api.get<DashboardStats>('/finance/dashboard').then((r) => r.data),

  // Invoices
  getInvoices: () =>
    api.get<Invoice[]>('/finance/invoices').then((r) => r.data),

  getInvoiceDetail: (id: string) =>
    api.get<Invoice>('/finance/invoices/detail', { params: { id } }).then((r) => r.data),

  createInvoice: (data: Partial<Invoice>) =>
    api.post<Invoice>('/finance/invoices', data).then((r) => r.data),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    api.post<Invoice>('/finance/invoices', { id, ...data }).then((r) => r.data),

  deleteInvoice: (id: string) =>
    api.post('/finance/invoices', { id, _action: 'delete' }).then((r) => r.data),

  updateInvoiceStatus: (id: string, status: Invoice['status']) =>
    api.post('/finance/invoices/status', { id, status }).then((r) => r.data),

  recordPayment: (id: string, amount: number) =>
    api.post('/finance/invoices/payment', { id, amount }).then((r) => r.data),

  // Expenses
  getExpenses: () =>
    api.get<Expense[]>('/finance/expenses').then((r) => r.data),

  createExpense: (data: Partial<Expense>) =>
    api.post<Expense>('/finance/expenses', data).then((r) => r.data),

  updateExpense: (id: string, data: Partial<Expense>) =>
    api.post<Expense>('/finance/expenses', { id, ...data }).then((r) => r.data),

  deleteExpense: (id: string) =>
    api.post('/finance/expenses', { id, _action: 'delete' }).then((r) => r.data),

  // Expense Categories
  getExpenseCategories: () =>
    api.get<ExpenseCategory[]>('/finance/expense-categories').then((r) => r.data),

  createExpenseCategory: (data: Partial<ExpenseCategory>) =>
    api.post<ExpenseCategory>('/finance/expense-categories', data).then((r) => r.data),

  // P&L
  getPL: (period: string) =>
    api.get<PLReport>('/finance/pl', { params: { period } }).then((r) => r.data),

  // Tax
  getTaxConfig: () =>
    api.get<TaxConfig[]>('/finance/tax/config').then((r) => r.data),

  createTaxConfig: (data: TaxConfig) =>
    api.post<TaxConfig>('/finance/tax/config', data).then((r) => r.data),

  getTaxReport: () =>
    api.get<TaxReport>('/finance/tax/report').then((r) => r.data),
};
