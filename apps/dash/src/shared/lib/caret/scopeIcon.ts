/** Convert a tool's scope label to a single emoji for compact UIs. */
export function scopeEmoji(scope: 'read' | 'write'): string {
  return scope === 'read' ? '👀' : '✏️';
}
