// Pure helpers for color token mapping based on POS domain states.

/** Return a Tailwind color token for a generic status value. */
export function statusColor(
  status: 'success' | 'warning' | 'error' | 'info' | 'neutral',
): string {
  const map: Record<typeof status, string> = {
    success: 'green',
    warning: 'yellow',
    error: 'red',
    info: 'blue',
    neutral: 'gray',
  };
  return map[status];
}

/** Return a color token for a numeric variance (positive=green, negative=red, zero=gray). */
export function varianceColor(variance: number): 'green' | 'red' | 'gray' {
  if (!Number.isFinite(variance) || variance === 0) return 'gray';
  return variance > 0 ? 'green' : 'red';
}

/** Return a color token for stock level. */
export function stockLevelColor(qty: number, threshold: number): 'red' | 'yellow' | 'green' {
  if (qty <= 0) return 'red';
  if (qty <= threshold) return 'yellow';
  return 'green';
}

/** Return a color token for a KDS timer urgency (seconds elapsed). */
export function kdsTimerColor(elapsedSeconds: number): 'green' | 'yellow' | 'red' {
  if (elapsedSeconds < 300) return 'green';
  if (elapsedSeconds < 600) return 'yellow';
  return 'red';
}

/** Return a color token for a loyalty tier based on points. */
export function loyaltyTierColor(points: number): 'gray' | 'yellow' | 'orange' | 'purple' {
  if (points >= 5000) return 'purple';
  if (points >= 2000) return 'orange';
  if (points >= 500) return 'yellow';
  return 'gray';
}
