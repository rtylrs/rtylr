// Compensation analysis helpers for the recruit admin panel.

export type CompensationType = 'salary' | 'hourly' | 'contract';

export interface CompensationBand {
  role: string;
  level: string;
  minSalary: number;
  midSalary: number;
  maxSalary: number;
  currency: string;
}

/**
 * bandRange returns the spread between max and min salary in the band.
 */
export function bandRange(band: CompensationBand): number {
  return band.maxSalary - band.minSalary;
}

/**
 * compaRatio returns the ratio of an offered salary to the band midpoint.
 */
export function compaRatio(offeredSalary: number, band: CompensationBand): number {
  if (band.midSalary === 0) return 0;
  return Math.round((offeredSalary / band.midSalary) * 1000) / 1000;
}

/**
 * isWithinBand returns true when the offered salary falls within the band range.
 */
export function isWithinBand(offeredSalary: number, band: CompensationBand): boolean {
  return offeredSalary >= band.minSalary && offeredSalary <= band.maxSalary;
}

/**
 * bandPositionPct returns the position of a salary within the band as a percentage.
 */
export function bandPositionPct(offeredSalary: number, band: CompensationBand): number {
  const range = bandRange(band);
  if (range === 0) return 0;
  const clamped = Math.max(band.minSalary, Math.min(band.maxSalary, offeredSalary));
  return Math.round(((clamped - band.minSalary) / range) * 1000) / 10;
}

/**
 * compensationTypeLabel returns a display label for a compensation type.
 */
export function compensationTypeLabel(type: CompensationType | string): string {
  const labels: Record<string, string> = {
    salary: 'Annual Salary',
    hourly: 'Hourly Rate',
    contract: 'Contract',
  };
  return labels[type] ?? type;
}

/**
 * formatSalary formats a salary value with currency and K notation for large amounts.
 */
export function formatSalary(amount: number, currency = 'USD'): string {
  if (amount >= 1000) {
    return `${currency} ${Math.round(amount / 1000)}K`;
  }
  return `${currency} ${amount}`;
}
