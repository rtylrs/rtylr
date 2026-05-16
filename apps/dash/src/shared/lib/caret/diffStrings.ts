/**
 * Compute the longest common prefix length between two strings. Used to
 * minimize re-renders when streaming text replaces earlier content.
 */
export function commonPrefixLength(a: string, b: string): number {
  const max = Math.min(a.length, b.length);
  let i = 0;
  while (i < max && a.charCodeAt(i) === b.charCodeAt(i)) i++;
  return i;
}
