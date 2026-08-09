/**
 * General formatting utilities for the Menu app.
 */

/** Format a number as a compact string (e.g. 1200 -> "1.2k"). */
export function formatCompact(value: number): string {
  if (Math.abs(value) >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (Math.abs(value) >= 1_000) return `${(value / 1_000).toFixed(1)}k`;
  return String(value);
}

/** Return the ordinal suffix for a number (e.g. 1 -> "st", 2 -> "nd"). */
export function ordinalSuffix(n: number): string {
  const abs = Math.abs(n);
  const mod100 = abs % 100;
  const mod10 = abs % 10;
  if (mod100 >= 11 && mod100 <= 13) return 'th';
  if (mod10 === 1) return 'st';
  if (mod10 === 2) return 'nd';
  if (mod10 === 3) return 'rd';
  return 'th';
}

/** Format a number with its ordinal suffix (e.g. 1 -> "1st"). */
export function formatOrdinal(n: number): string {
  return `${n}${ordinalSuffix(n)}`;
}

/** Pluralise a word based on count (e.g. 1 "item" -> "1 item", 2 -> "2 items"). */
export function pluralise(count: number, singular: string, plural?: string): string {
  const word = count === 1 ? singular : (plural ?? `${singular}s`);
  return `${count} ${word}`;
}

/** Left-pad a string with a fill character to reach a target length. */
export function padStart(value: string, length: number, fill = ' '): string {
  return value.padStart(length, fill);
}

/** Right-pad a string with a fill character to reach a target length. */
export function padEnd(value: string, length: number, fill = ' '): string {
  return value.padEnd(length, fill);
}

/** Format bytes as a human-readable string (e.g. 1024 -> "1.0 KB"). */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}
