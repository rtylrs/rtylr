export type ReportPeriod = 'weekly' | 'monthly' | 'quarterly' | 'yearly';
export type ReportFormat = 'csv' | 'pdf' | 'xlsx';

export function reportPeriodLabel(period: ReportPeriod): string {
  const labels: Record<ReportPeriod, string> = {
    weekly: 'Weekly',
    monthly: 'Monthly',
    quarterly: 'Quarterly',
    yearly: 'Yearly',
  };
  return labels[period];
}

export function reportFormatLabel(format: ReportFormat): string {
  const labels: Record<ReportFormat, string> = {
    csv: 'CSV',
    pdf: 'PDF',
    xlsx: 'Excel',
  };
  return labels[format];
}

export function reportFormatMimeType(format: ReportFormat): string {
  const types: Record<ReportFormat, string> = {
    csv: 'text/csv',
    pdf: 'application/pdf',
    xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  };
  return types[format];
}

export function periodStartDate(referenceIso: string, period: ReportPeriod): string {
  const d = new Date(referenceIso);
  if (period === 'weekly') {
    const day = d.getDay();
    d.setDate(d.getDate() - day);
  } else if (period === 'monthly') {
    d.setDate(1);
  } else if (period === 'quarterly') {
    const month = d.getMonth();
    const quarterStart = month - (month % 3);
    d.setMonth(quarterStart, 1);
  } else if (period === 'yearly') {
    d.setMonth(0, 1);
  }
  return d.toISOString().slice(0, 10);
}

export function percentageChange(previous: number, current: number): number | null {
  if (previous === 0) return null;
  return ((current - previous) / Math.abs(previous)) * 100;
}

export function reportFilename(
  name: string,
  period: ReportPeriod,
  dateIso: string,
  format: ReportFormat
): string {
  const safeName = name.toLowerCase().replace(/\s+/g, '_');
  return `${safeName}_${period}_${dateIso}.${format}`;
}

export function groupAndCount<T>(records: T[], keyFn: (item: T) => string): Record<string, number> {
  const result: Record<string, number> = {};
  for (const record of records) {
    const key = keyFn(record);
    result[key] = (result[key] ?? 0) + 1;
  }
  return result;
}
