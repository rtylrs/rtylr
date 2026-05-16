/** Filter out null/undefined entries. */
export function compact<T>(items: (T | null | undefined)[]): T[] {
  const out: T[] = [];
  for (const x of items) if (x != null) out.push(x);
  return out;
}
