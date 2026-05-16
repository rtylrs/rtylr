/** Generate an opaque stream/request identifier. */
export function newStreamId(): string {
  const r = crypto?.getRandomValues?.(new Uint8Array(4)) ?? new Uint8Array(4);
  let hex = '';
  for (const byte of r) hex += byte.toString(16).padStart(2, '0');
  return `s-${Date.now().toString(36)}-${hex}`;
}
