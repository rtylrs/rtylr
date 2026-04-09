// Mock data generators used by hooks when the API is unavailable.
import type { LineItem } from '../components/LineItemRow';
import type { AccountType } from '../components/AccountTypeBadge';

export interface InvoiceDetail {
  id: string;
  client: string;
  clientEmail: string;
  amount: number;
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue' | 'Partial';
  issuedDate: string;
  dueDate: string;
  notes: string;
  items: LineItem[];
  history: Array<{ at: string; label: string }>;
  discount: number;
  paid: number;
}

export function buildMockInvoice(id: string): InvoiceDetail {
  const items: LineItem[] = [
    { id: 'l1', description: 'Strategy workshop (4 hr)', quantity: 4, unitPrice: 250, taxRate: 8.5 },
    { id: 'l2', description: 'Brand audit deliverable', quantity: 1, unitPrice: 1800, taxRate: 8.5 },
    { id: 'l3', description: 'Implementation support', quantity: 8, unitPrice: 125, taxRate: 0 },
  ];
  const subtotal = items.reduce((s, it) => s + it.quantity * it.unitPrice, 0);
  const tax = items.reduce((s, it) => s + it.quantity * it.unitPrice * (it.taxRate / 100), 0);
  return {
    id,
    client: 'Acme Corporation',
    clientEmail: 'billing@acme.com',
    amount: subtotal + tax,
    status: 'Sent',
    issuedDate: '2026-03-15',
    dueDate: '2026-04-15',
    notes: 'Thank you for your business. Payment due within 30 days.',
    items,
    history: [
      { at: '2026-03-15 09:12', label: 'Invoice created' },
      { at: '2026-03-15 09:20', label: 'Sent to billing@acme.com' },
      { at: '2026-03-18 14:03', label: 'Viewed by recipient' },
    ],
    discount: 0,
    paid: 0,
  };
}

export interface RecurringInvoice {
  id: string;
  client: string;
  amount: number;
  cadence: 'monthly' | 'quarterly' | 'yearly' | 'weekly';
  nextDate: string;
  status: 'active' | 'paused';
  lastIssued: string;
}

export const MOCK_RECURRING: RecurringInvoice[] = [
  { id: 'r-01', client: 'Globex Industries', amount: 2400, cadence: 'monthly', nextDate: '2026-05-01', status: 'active', lastIssued: '2026-04-01' },
  { id: 'r-02', client: 'Wayne Enterprises', amount: 9800, cadence: 'quarterly', nextDate: '2026-07-01', status: 'active', lastIssued: '2026-04-01' },
  { id: 'r-03', client: 'Initech Ltd.', amount: 1250, cadence: 'monthly', nextDate: '2026-04-15', status: 'paused', lastIssued: '2026-03-15' },
  { id: 'r-04', client: 'Pied Piper', amount: 580, cadence: 'weekly', nextDate: '2026-04-12', status: 'active', lastIssued: '2026-04-05' },
  { id: 'r-05', client: 'Umbrella LLC', amount: 12000, cadence: 'yearly', nextDate: '2027-01-15', status: 'active', lastIssued: '2026-01-15' },
];

export interface Quote {
  id: string;
  client: string;
  amount: number;
  status: 'Draft' | 'Sent' | 'Accepted' | 'Declined' | 'Expired';
  issuedDate: string;
  expiresAt: string;
}

export const MOCK_QUOTES: Quote[] = [
  { id: 'Q-2026-001', client: 'Stark Industries', amount: 28500, status: 'Sent', issuedDate: '2026-04-01', expiresAt: '2026-05-01' },
  { id: 'Q-2026-002', client: 'Wonka Inc', amount: 4200, status: 'Accepted', issuedDate: '2026-03-28', expiresAt: '2026-04-28' },
  { id: 'Q-2026-003', client: 'Globex Industries', amount: 7800, status: 'Draft', issuedDate: '2026-04-05', expiresAt: '2026-05-05' },
  { id: 'Q-2026-004', client: 'Initech Ltd.', amount: 15600, status: 'Declined', issuedDate: '2026-03-20', expiresAt: '2026-04-20' },
  { id: 'Q-2026-005', client: 'Umbrella LLC', amount: 9250, status: 'Expired', issuedDate: '2026-02-15', expiresAt: '2026-03-15' },
];

export interface ExpenseCategoryBudget {
  id: string;
  name: string;
  color: string;
  budget: number;
  spent: number;
}

export const MOCK_CATEGORIES: ExpenseCategoryBudget[] = [
  { id: 'cat-1', name: 'Software & SaaS', color: '#3b82f6', budget: 2500, spent: 1820 },
  { id: 'cat-2', name: 'Travel', color: '#8b5cf6', budget: 4000, spent: 2870 },
  { id: 'cat-3', name: 'Marketing', color: '#ec4899', budget: 6000, spent: 5200 },
  { id: 'cat-4', name: 'Office Supplies', color: '#f59e0b', budget: 800, spent: 320 },
  { id: 'cat-5', name: 'Meals & Entertainment', color: '#ef4444', budget: 1500, spent: 1640 },
  { id: 'cat-6', name: 'Utilities', color: '#10b981', budget: 900, spent: 720 },
  { id: 'cat-7', name: 'Contractors', color: '#06b6d4', budget: 8000, spent: 6500 },
];

export interface Bill {
  id: string;
  vendor: string;
  amount: number;
  dueDate: string;
  status: 'unpaid' | 'paid' | 'overdue' | 'scheduled';
  category: string;
}

export const MOCK_BILLS: Bill[] = [
  { id: 'B-1001', vendor: 'AWS', amount: 4820, dueDate: '2026-04-25', status: 'unpaid', category: 'Software & SaaS' },
  { id: 'B-1002', vendor: 'Figma', amount: 420, dueDate: '2026-04-15', status: 'unpaid', category: 'Software & SaaS' },
  { id: 'B-1003', vendor: 'Delta Airlines', amount: 1280, dueDate: '2026-04-02', status: 'paid', category: 'Travel' },
  { id: 'B-1004', vendor: 'Slack', amount: 320, dueDate: '2026-03-30', status: 'overdue', category: 'Software & SaaS' },
  { id: 'B-1005', vendor: 'Adobe CC', amount: 599, dueDate: '2026-05-01', status: 'scheduled', category: 'Software & SaaS' },
  { id: 'B-1006', vendor: 'WeWork', amount: 3200, dueDate: '2026-04-30', status: 'unpaid', category: 'Office' },
  { id: 'B-1007', vendor: 'Google Workspace', amount: 218, dueDate: '2026-04-10', status: 'paid', category: 'Software & SaaS' },
];

export interface BankAccount {
  id: string;
  institution: string;
  mask: string;
  type: 'checking' | 'savings' | 'credit';
  balance: number;
  currency: string;
}

export interface BankTransaction {
  id: string;
  accountId: string;
  date: string;
  description: string;
  amount: number;
  category: string;
  matched: boolean;
}

export const MOCK_BANK_ACCOUNTS: BankAccount[] = [
  { id: 'acct-1', institution: 'Chase', mask: '4201', type: 'checking', balance: 128450.32, currency: 'USD' },
  { id: 'acct-2', institution: 'Chase', mask: '9974', type: 'savings', balance: 340210.12, currency: 'USD' },
  { id: 'acct-3', institution: 'Amex', mask: '1008', type: 'credit', balance: -8720.55, currency: 'USD' },
];

export const MOCK_BANK_TX: BankTransaction[] = [
  { id: 'tx-1', accountId: 'acct-1', date: '2026-04-08', description: 'Wire from Acme Corporation', amount: 14820, category: 'Revenue', matched: false },
  { id: 'tx-2', accountId: 'acct-1', date: '2026-04-07', description: 'AWS monthly bill', amount: -4820, category: 'Software', matched: true },
  { id: 'tx-3', accountId: 'acct-1', date: '2026-04-06', description: 'Payroll run', amount: -42800, category: 'Payroll', matched: true },
  { id: 'tx-4', accountId: 'acct-3', date: '2026-04-05', description: 'Delta Airlines flight', amount: -1280, category: 'Travel', matched: true },
  { id: 'tx-5', accountId: 'acct-1', date: '2026-04-04', description: 'Stripe payout', amount: 8920, category: 'Revenue', matched: false },
  { id: 'tx-6', accountId: 'acct-3', date: '2026-04-03', description: 'Figma subscription', amount: -420, category: 'Software', matched: false },
  { id: 'tx-7', accountId: 'acct-1', date: '2026-04-02', description: 'Wire from Globex Industries', amount: 2400, category: 'Revenue', matched: false },
  { id: 'tx-8', accountId: 'acct-1', date: '2026-04-01', description: 'Office rent', amount: -6800, category: 'Rent', matched: true },
];

export interface BalanceSheetRow {
  label: string;
  amount: number;
  type: AccountType;
}

export interface BalanceSheet {
  assets: BalanceSheetRow[];
  liabilities: BalanceSheetRow[];
  equity: BalanceSheetRow[];
}

export const MOCK_BALANCE_SHEET: BalanceSheet = {
  assets: [
    { label: 'Cash - Chase Checking', amount: 128450.32, type: 'asset' },
    { label: 'Cash - Chase Savings', amount: 340210.12, type: 'asset' },
    { label: 'Accounts Receivable', amount: 82390.55, type: 'asset' },
    { label: 'Prepaid Expenses', amount: 12800.0, type: 'asset' },
    { label: 'Equipment', amount: 45000.0, type: 'asset' },
  ],
  liabilities: [
    { label: 'Credit Card (Amex)', amount: 8720.55, type: 'liability' },
    { label: 'Accounts Payable', amount: 18340.25, type: 'liability' },
    { label: 'Accrued Payroll', amount: 28500.0, type: 'liability' },
    { label: 'Deferred Revenue', amount: 32000.0, type: 'liability' },
  ],
  equity: [
    { label: "Owner's Equity", amount: 420000.0, type: 'equity' },
    { label: 'Retained Earnings', amount: 101290.19, type: 'equity' },
  ],
};

export interface CashFlowMonth {
  month: string;
  inflow: number;
  outflow: number;
}

export const MOCK_CASH_FLOW: CashFlowMonth[] = [
  { month: 'Oct', inflow: 82000, outflow: 54200 },
  { month: 'Nov', inflow: 95400, outflow: 58900 },
  { month: 'Dec', inflow: 118000, outflow: 71200 },
  { month: 'Jan', inflow: 86500, outflow: 62400 },
  { month: 'Feb', inflow: 94300, outflow: 65700 },
  { month: 'Mar', inflow: 102800, outflow: 69200 },
  { month: 'Apr', inflow: 76400, outflow: 42300 },
];

export interface JournalEntry {
  id: string;
  date: string;
  memo: string;
  lines: Array<{ account: string; debit: number; credit: number }>;
}

export const MOCK_JOURNAL: JournalEntry[] = [
  {
    id: 'je-001',
    date: '2026-04-08',
    memo: 'Payment received from Acme',
    lines: [
      { account: 'Cash - Chase', debit: 14820, credit: 0 },
      { account: 'Accounts Receivable', debit: 0, credit: 14820 },
    ],
  },
  {
    id: 'je-002',
    date: '2026-04-07',
    memo: 'AWS monthly bill',
    lines: [
      { account: 'Software Expense', debit: 4820, credit: 0 },
      { account: 'Cash - Chase', debit: 0, credit: 4820 },
    ],
  },
  {
    id: 'je-003',
    date: '2026-04-06',
    memo: 'Payroll',
    lines: [
      { account: 'Salary Expense', debit: 42800, credit: 0 },
      { account: 'Cash - Chase', debit: 0, credit: 42800 },
    ],
  },
  {
    id: 'je-004',
    date: '2026-04-02',
    memo: 'Invoice issued to Globex',
    lines: [
      { account: 'Accounts Receivable', debit: 2400, credit: 0 },
      { account: 'Revenue', debit: 0, credit: 2400 },
    ],
  },
];

export interface Account {
  id: string;
  code: string;
  name: string;
  type: AccountType;
  balance: number;
}

export const MOCK_ACCOUNTS: Account[] = [
  { id: 'a-100', code: '1000', name: 'Cash - Chase Checking', type: 'asset', balance: 128450.32 },
  { id: 'a-101', code: '1010', name: 'Cash - Chase Savings', type: 'asset', balance: 340210.12 },
  { id: 'a-102', code: '1200', name: 'Accounts Receivable', type: 'asset', balance: 82390.55 },
  { id: 'a-103', code: '1500', name: 'Equipment', type: 'asset', balance: 45000.0 },
  { id: 'a-200', code: '2000', name: 'Accounts Payable', type: 'liability', balance: 18340.25 },
  { id: 'a-201', code: '2100', name: 'Credit Card - Amex', type: 'liability', balance: 8720.55 },
  { id: 'a-202', code: '2300', name: 'Deferred Revenue', type: 'liability', balance: 32000.0 },
  { id: 'a-300', code: '3000', name: "Owner's Equity", type: 'equity', balance: 420000.0 },
  { id: 'a-301', code: '3100', name: 'Retained Earnings', type: 'equity', balance: 101290.19 },
  { id: 'a-400', code: '4000', name: 'Product Revenue', type: 'revenue', balance: 482000.0 },
  { id: 'a-401', code: '4100', name: 'Services Revenue', type: 'revenue', balance: 218400.0 },
  { id: 'a-500', code: '5000', name: 'Software Expense', type: 'expense', balance: 28450.0 },
  { id: 'a-501', code: '5100', name: 'Salary Expense', type: 'expense', balance: 342000.0 },
  { id: 'a-502', code: '5200', name: 'Travel Expense', type: 'expense', balance: 18420.0 },
  { id: 'a-503', code: '5300', name: 'Marketing Expense', type: 'expense', balance: 48200.0 },
];

export interface ExpenseDetail {
  id: string;
  date: string;
  description: string;
  category: string;
  amount: number;
  submitter: string;
  approval: 'pending' | 'approved' | 'rejected';
  notes: string;
  receiptUrl: string | null;
}

export function buildMockExpense(id: string): ExpenseDetail {
  return {
    id,
    date: '2026-04-05',
    description: 'Client dinner - Stark Industries',
    category: 'Meals & Entertainment',
    amount: 284.32,
    submitter: 'Jane Doe',
    approval: 'pending',
    notes: 'Dinner meeting to discuss Q2 proposal.',
    receiptUrl: null,
  };
}
