import { toFiniteNumber } from '@voxire/api';

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(toFiniteNumber(value));
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(toFiniteNumber(value));
}

// Minutes as hours to one decimal, with the language's separator but Latin
// digits like the rest of the dashboard.
export function formatHours(minutes: number, language: string): string {
  const hours = toFiniteNumber(minutes) / 60;
  const options: Intl.NumberFormatOptions = { maximumFractionDigits: 1 };
  try {
    return new Intl.NumberFormat(`${language}-u-nu-latn`, options).format(hours);
  } catch {
    return new Intl.NumberFormat('en-US', options).format(hours);
  }
}

export function formatPercent(value: number): string {
  const safeValue = toFiniteNumber(value);
  return `${safeValue >= 0 ? '+' : ''}${safeValue.toFixed(1)}%`;
}

export function timeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diffMs = now - then;
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return 'Just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 30) return `${diffDay}d ago`;
  const diffMo = Math.floor(diffDay / 30);
  return `${diffMo}mo ago`;
}
