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
  const csv = rows.map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n');
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
    ...invoices.map((i) => [i.id, i.client, i.issuedDate, i.dueDate, i.amount.toFixed(2), i.status, 'Finance']),
  ];
  downloadCsv('quickbooks_invoices.csv', rows);
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
