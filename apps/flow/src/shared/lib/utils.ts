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
  const now = new Date();
  const diff = d.getTime() - now.getTime();
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24));

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
