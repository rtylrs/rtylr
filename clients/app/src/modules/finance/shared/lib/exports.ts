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

function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows
    .map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(','))
    .join('\n');
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
