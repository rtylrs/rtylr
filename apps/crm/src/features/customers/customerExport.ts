import { getCustomerTier, type Tier } from '../../shared/lib/tiers';
import type { Customer } from '../../shared/schemas/customers';

interface CsvCell {
  value: string;
  protectFormula?: boolean;
}

function escapeCsvCell(cell: CsvCell): string {
  const safeValue =
    cell.protectFormula !== false && /^[\t\r\n ]*[=+\-@]/.test(cell.value)
      ? `'${cell.value}`
      : cell.value;
  return `"${safeValue.replace(/"/g, '""')}"`;
}

/** Serializes customer data for spreadsheet-compatible CSV downloads. */
export function customersToCsv(customers: Customer[], thresholds: Record<Tier, number>): string {
  const header: CsvCell[] = ['Name', 'Email', 'Phone', 'Spent', 'Visits', 'Points', 'Tier'].map(
    (value) => ({ value, protectFormula: false })
  );
  const rows: CsvCell[][] = customers.map((customer) => [
    { value: customer.name },
    { value: customer.email },
    { value: customer.phone },
    { value: String(customer.totalSpent), protectFormula: false },
    { value: String(customer.visits), protectFormula: false },
    { value: String(customer.loyaltyPoints), protectFormula: false },
    {
      value: getCustomerTier(customer.loyaltyPoints, customer.loyaltyTier, thresholds),
      protectFormula: false,
    },
  ]);

  return [header, ...rows].map((row) => row.map(escapeCsvCell).join(',')).join('\n');
}
