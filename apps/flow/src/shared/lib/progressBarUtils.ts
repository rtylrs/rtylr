/** Returns the width percentage string for a progress bar. */
export function progressBarWidth(percent: number): string {
  const clamped = Math.min(100, Math.max(0, percent));
  return `${clamped}%`;
}

/** Returns a color class based on progress percentage. */
export function progressColor(percent: number): 'success' | 'warning' | 'danger' | 'default' {
  if (percent >= 80) return 'success';
  if (percent >= 50) return 'warning';
  if (percent >= 1) return 'danger';
  return 'default';
}

/** Returns a health color based on project health status. */
export function healthColor(
  health: 'on_track' | 'at_risk' | 'off_track'
): 'green' | 'yellow' | 'red' {
  const map: Record<'on_track' | 'at_risk' | 'off_track', 'green' | 'yellow' | 'red'> = {
    on_track: 'green',
    at_risk: 'yellow',
    off_track: 'red',
  };
  return map[health];
}

/** Returns the label to show alongside a progress bar. */
export function progressLabel(completed: number, total: number): string {
  if (total === 0) return '0 / 0';
  return `${completed} / ${total}`;
}

/** Returns a rounded percentage, or 0 if total is 0. */
export function progressPercent(completed: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((completed / total) * 100);
}

/** Returns a description of progress in words. */
export function progressDescription(percent: number): string {
  if (percent === 0) return 'Not started';
  if (percent < 25) return 'Just started';
  if (percent < 50) return 'In progress';
  if (percent < 75) return 'More than halfway';
  if (percent < 100) return 'Almost done';
  return 'Complete';
}
