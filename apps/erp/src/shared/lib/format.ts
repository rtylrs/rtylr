import { toFiniteNumber } from '@voxire/api';

export function formatCurrency(value: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  }).format(toFiniteNumber(value));
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(toFiniteNumber(value));
}

export function formatDate(dateString: string): string {
  try {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  } catch {
    return dateString;
  }
}

export function formatDateTime(dateString: string): string {
  try {
    return new Date(dateString).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateString;
  }
}

export function timeAgo(dateString: string): string {
  try {
    const now = new Date().getTime();
    const then = new Date(dateString).getTime();
    const diff = Math.floor((now - then) / 1000);
    if (diff < 60) return 'just now';
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`;
    return formatDate(dateString);
  } catch {
    return dateString;
  }
}

export function daysUntil(dateString: string): number {
  const now = new Date();
  const then = new Date(dateString);
  return Math.ceil((then.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}
