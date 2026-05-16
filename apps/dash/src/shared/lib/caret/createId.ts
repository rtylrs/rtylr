/** Simple sortable id mixing high-res time + random suffix. */
let counter = 0;
export function createId(prefix = 'id'): string {
  const t = Date.now().toString(36);
  const c = (counter = (counter + 1) % 9999).toString(36).padStart(3, '0');
  return `${prefix}-${t}-${c}`;
}
