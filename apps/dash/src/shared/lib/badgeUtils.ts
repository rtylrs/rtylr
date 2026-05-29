/**
 * Badge display utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type BadgeColor =
  | 'blue'
  | 'green'
  | 'yellow'
  | 'red'
  | 'purple'
  | 'gray'
  | 'orange'
  | 'teal';

export interface BadgeConfig {
  label: string;
  color: BadgeColor;
}

/** Return a count badge label, capping at a max display value. */
export function countBadge(count: number, max = 99): string {
  if (count <= 0) return '0';
  return count > max ? `${max}+` : String(count);
}

/** Return true when the count warrants showing a badge. */
export function shouldShowBadge(count: number): boolean {
  return count > 0;
}

/** Abbreviate a long label to a max length with an ellipsis. */
export function truncateBadgeLabel(label: string, maxLen = 20): string {
  if (label.length <= maxLen) return label;
  return `${label.slice(0, maxLen - 1)}...`;
}

/** Return a color from a palette based on a hash of the given string. */
export function colorForString(str: string): BadgeColor {
  const colors: BadgeColor[] = [
    'blue',
    'green',
    'yellow',
    'red',
    'purple',
    'gray',
    'orange',
    'teal',
  ];
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 31 + str.charCodeAt(i)) >>> 0;
  }
  return colors[hash % colors.length]!;
}

/** Map a numeric score (0-100) to a badge color from red to green. */
export function scoreColor(score: number): BadgeColor {
  if (score >= 80) return 'green';
  if (score >= 60) return 'teal';
  if (score >= 40) return 'yellow';
  if (score >= 20) return 'orange';
  return 'red';
}

/** Build a BadgeConfig from a label and optional explicit color. */
export function makeBadge(label: string, color?: BadgeColor): BadgeConfig {
  return { label, color: color ?? colorForString(label) };
}
