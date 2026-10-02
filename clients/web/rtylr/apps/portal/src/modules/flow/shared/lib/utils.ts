import { daysUntil } from './dateUtils';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { TFunction } from 'i18next';

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

/**
 * An absolute, localised day label such as "Sep 7".
 *
 * For ranges and start dates, not deadlines: a cycle that began last week is
 * running, not overdue, and the relative due-date vocabulary said otherwise
 * (rtylr#1221).
 */
export function formatShortDate(date: string, locale = 'en-US'): string {
  if (!date || String(date).trim() === '') return '-';
  const d = new Date(date);
  try {
    return d.toLocaleDateString(locale, { month: 'short', day: 'numeric' });
  } catch {
    // The locale comes from translation resources; if it is ever not a valid
    // BCP 47 tag, render in the default rather than take the screen down.
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }
}

/**
 * A relative label for a due date: "Today", "Tomorrow", "Yesterday",
 * "3d overdue", or the short date further out.
 *
 * Calendar days apart, not a millisecond gap rounded up: the old arithmetic
 * subtracted a local `new Date()` from a bare YYYY-MM-DD, which parses as UTC
 * midnight, so the two sides were in different frames. At UTC+3 a task due
 * today, read at 02:00, came out one day in the future and the board labelled
 * it "Tomorrow" (#1192).
 */
export function formatDueDate(date: string, t: TFunction): string {
  if (!date || String(date).trim() === '') return '-';
  const days = daysUntil(date);

  if (days === 0) return t('flowCommon.dueDateLabel.today');
  if (days === 1) return t('flowCommon.dueDateLabel.tomorrow');
  if (days === -1) return t('flowCommon.dueDateLabel.yesterday');
  if (days < 0) return t('flowCommon.dueDateLabel.overdue', { days: Math.abs(days) });

  return formatShortDate(date, t('flowCommon.dateLocale'));
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
