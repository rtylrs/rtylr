import { daysUntil } from './dateUtils';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

export function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

export function formatDate(date: string): string {
  if (!date || String(date).trim() === '') return '-';
  const d = new Date(date);

  // Calendar days apart, not a millisecond gap rounded up.
  //
  // The old arithmetic subtracted a local `new Date()` from a bare YYYY-MM-DD,
  // which parses as UTC midnight, so the two sides were in different frames.
  // At UTC+3 a task due *today*, read at 02:00, came out one day in the future
  // and the board labelled it "Tomorrow". That is the symptom the design
  // review reported and attributed to `todayIso`; #1192 fixed that helper,
  // while this formatter kept its own broken sum.
  //
  // `daysUntil` compares two ISO days through the same frame, so the answer is
  // a whole number of calendar days and does not depend on the hour.
  const days = daysUntil(date);

  if (days === 0) return 'Today';
  if (days === 1) return 'Tomorrow';
  if (days === -1) return 'Yesterday';
  if (days < 0) return `${Math.abs(days)}d overdue`;

  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;
  const nameEQ = `${name}=`;
  const parts = document.cookie.split('; ');
  for (const part of parts) {
    if (part.startsWith(nameEQ)) {
      return decodeURIComponent(part.substring(nameEQ.length));
    }
  }
  return null;
}
