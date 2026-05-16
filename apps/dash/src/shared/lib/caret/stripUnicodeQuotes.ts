/** Replace curly quotes with ASCII so the prompt round-trips through APIs. */
export function asciiQuotes(s: string): string {
  return s.replace(/[‘’]/g, "'").replace(/[“”]/g, '"');
}
