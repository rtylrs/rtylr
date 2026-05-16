/** Escape HTML entities. Identical to the @voxire/caret helper, re-exposed
 *  for dash-app callers that don't want the full package dependency. */
export function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
