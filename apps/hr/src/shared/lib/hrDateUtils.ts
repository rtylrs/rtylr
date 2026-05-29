export function formatDate(
  dateIso: string,
  format: 'short' | 'long' | 'numeric' = 'short'
): string {
  const d = new Date(dateIso);
  if (isNaN(d.getTime())) return '';
  if (format === 'numeric') return d.toLocaleDateString('en-US');
  if (format === 'long')
    return d.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

export function daysBetween(startIso: string, endIso: string): number {
  const msPerDay = 1000 * 60 * 60 * 24;
  return Math.round((new Date(endIso).getTime() - new Date(startIso).getTime()) / msPerDay);
}

export function addDays(dateIso: string, days: number): string {
  const d = new Date(dateIso);
  d.setDate(d.getDate() + days);
  return d.toISOString().slice(0, 10);
}

export function startOfMonth(dateIso: string): string {
  const d = new Date(dateIso);
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), 1)).toISOString().slice(0, 10);
}

export function endOfMonth(dateIso: string): string {
  const d = new Date(dateIso);
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + 1, 0)).toISOString().slice(0, 10);
}

export function isWeekend(dateIso: string): boolean {
  const day = new Date(dateIso).getDay();
  return day === 0 || day === 6;
}

export function isToday(dateIso: string): boolean {
  const today = new Date().toISOString().slice(0, 10);
  return dateIso === today;
}

export function isFutureDate(dateIso: string): boolean {
  return new Date(dateIso) > new Date();
}

export function isPastDate(dateIso: string): boolean {
  return new Date(dateIso) < new Date();
}

export function yearsOld(birthDateIso: string): number {
  const today = new Date();
  const birth = new Date(birthDateIso);
  let years = today.getFullYear() - birth.getFullYear();
  const monthDiff = today.getMonth() - birth.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    years--;
  }
  return years;
}
