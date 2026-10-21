// Simple CSV exporters for QuickBooks and Xero.
import { toast } from 'sonner';

export interface InvoiceExport {
  id: string;
  client: string;
  amount: number;
  issuedDate: string;
  dueDate: string;
  status: string;
}

/**
 * Quotes one cell. A cell a spreadsheet would read as a formula (it starts
 * with =, +, -, @, tab or carriage return and is not a plain number) is
 * prefixed with an apostrophe, so a vendor named "=HYPERLINK(...)" stays text
 * while a loss of "-12.00" stays a number.
 */
function csvCell(value: unknown): string {
  let text = String(value);
  const isNumber = /^[+-]?\d+(\.\d+)?$/.test(text);
  if (!isNumber && /^[=+\-@\t\r]/.test(text)) text = `'${text}`;
  return `"${text.replace(/"/g, '""')}"`;
}

function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows.map((r) => r.map(csvCell).join(',')).join('\n');
  const blob = new Blob([csv], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
  toast.success(`Downloaded ${filename}`);
}

export function exportQuickbooksInvoices(invoices: InvoiceExport[]) {
  // QuickBooks IIF-style CSV (simplified).
  const header = ['InvoiceNo', 'Customer', 'InvoiceDate', 'DueDate', 'Amount', 'Memo', 'Class'];
  const rows = [
    header,
    ...invoices.map((i) => [
      i.id,
      i.client,
      i.issuedDate,
      i.dueDate,
      i.amount.toFixed(2),
      i.status,
      'Finance',
    ]),
  ];
  downloadCsv('quickbooks_invoices.csv', rows);
}

export interface PLExportRow {
  period: string;
  revenue: number;
  cogs: number;
  grossProfit: number;
  operatingExpenses: number;
  netProfit: number;
}

export function exportProfitLossCsv(row: PLExportRow): void {
  const header = ['Period', 'Revenue', 'COGS', 'Gross Profit', 'Operating Expenses', 'Net Profit'];
  const rows = [
    header,
    [
      row.period,
      row.revenue.toFixed(2),
      row.cogs.toFixed(2),
      row.grossProfit.toFixed(2),
      row.operatingExpenses.toFixed(2),
      row.netProfit.toFixed(2),
    ],
  ];
  downloadCsv(`profit_loss_${row.period.replace(/\s+/g, '_').toLowerCase()}.csv`, rows);
}

export interface TaxExportRow {
  period: string;
  collected: number;
  owed: number;
  paid: number;
}

export function exportTaxSummaryCsv(rows: TaxExportRow[]): void {
  const header = ['Period', 'Collected', 'Owed', 'Paid'];
  const data = [
    header,
    ...rows.map((r) => [r.period, r.collected.toFixed(2), r.owed.toFixed(2), r.paid.toFixed(2)]),
  ];
  downloadCsv('tax_summary.csv', data);
}

export function exportXeroInvoices(invoices: InvoiceExport[]) {
  // Xero invoice template CSV header (simplified).
  const header = [
    '*ContactName',
    '*InvoiceNumber',
    '*InvoiceDate',
    '*DueDate',
    '*Description',
    '*Quantity',
    '*UnitAmount',
    '*AccountCode',
    'TaxType',
  ];
  const rows = [
    header,
    ...invoices.map((i) => [
      i.client,
      i.id,
      i.issuedDate,
      i.dueDate,
      `Invoice ${i.id}`,
      '1',
      i.amount.toFixed(2),
      '200',
      'Tax on Sales',
    ]),
  ];
  downloadCsv('xero_invoices.csv', rows);
}

export interface ExpenseExportRow {
  date: string;
  description: string;
  vendor: string;
  category: string;
  paymentMethod: string;
  status: string;
  amount: number;
  currency: string;
  receiptUrl: string;
}

export function exportExpensesCsv(filename: string, rows: ExpenseExportRow[]): void {
  downloadCsv(filename, [
    [
      'Date',
      'Description',
      'Vendor',
      'Category',
      'Payment method',
      'Status',
      'Amount',
      'Currency',
      'Receipt',
    ],
    ...rows.map((r) => [
      r.date,
      r.description,
      r.vendor,
      r.category,
      r.paymentMethod,
      r.status,
      r.amount.toFixed(2),
      r.currency,
      r.receiptUrl,
    ]),
  ]);
}
