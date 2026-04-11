// Date helpers used across the finance app.

export function formatDate(d: Date | string, opts?: Intl.DateTimeFormatOptions): string {
  const date = typeof d === 'string' ? new Date(d) : d;
  if (isNaN(date.getTime())) return '-';
  return date.toLocaleDateString(
    'en-US',
    opts ?? { month: 'short', day: 'numeric', year: 'numeric' }
  );
}

export function formatDateShort(d: Date | string): string {
  return formatDate(d, { month: 'short', day: 'numeric' });
}

export function daysBetween(a: Date | string, b: Date | string): number {
  const d1 = typeof a === 'string' ? new Date(a) : a;
  const d2 = typeof b === 'string' ? new Date(b) : b;
  return Math.floor((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24));
}

export function addDays(d: Date | string, days: number): Date {
  const date = typeof d === 'string' ? new Date(d) : new Date(d);
  date.setDate(date.getDate() + days);
  return date;
}

export function monthKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

export function monthLabel(d: Date | string): string {
  const date = typeof d === 'string' ? new Date(d) : d;
  return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
}

export function fiscalYearStartFromDate(d: Date, fiscalStartMonth = 1): Date {
  const year = d.getMonth() + 1 >= fiscalStartMonth ? d.getFullYear() : d.getFullYear() - 1;
  return new Date(year, fiscalStartMonth - 1, 1);
}

export function quarterOf(d: Date): number {
  return Math.floor(d.getMonth() / 3) + 1;
}
