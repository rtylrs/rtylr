/** Walk a dotted path through an object; returns undefined if anything is missing. */
export function safeGet<T = unknown>(obj: unknown, path: string): T | undefined {
  if (!path) return obj as T;
  const parts = path.split('.');
  let cur: unknown = obj;
  for (const p of parts) {
    if (cur == null || typeof cur !== 'object') return undefined;
    cur = (cur as Record<string, unknown>)[p];
  }
  return cur as T;
}
