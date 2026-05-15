/** Pluralize a noun by count with optional irregular plural. */
export function pluralize(count: number, singular: string, plural?: string): string {
  if (count === 1) return `1 ${singular}`;
  const word = plural ?? `${singular}s`;
  return `${count.toLocaleString()} ${word}`;
}
