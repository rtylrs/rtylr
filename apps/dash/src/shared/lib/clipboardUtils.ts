/**
 * Clipboard and copy-paste helpers for the Dash app.
 * All functions are pure with no side-effects where possible.
 */

/**
 * Copies a string to the system clipboard.
 * Returns true on success and false if the API is unavailable.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (!navigator?.clipboard?.writeText) return false;
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

/**
 * Reads text from the system clipboard.
 * Returns null if the API is unavailable or access is denied.
 */
export async function readFromClipboard(): Promise<string | null> {
  if (!navigator?.clipboard?.readText) return null;
  try {
    return await navigator.clipboard.readText();
  } catch {
    return null;
  }
}

/** Returns true if the Clipboard API is available in the current environment. */
export function isClipboardAvailable(): boolean {
  return typeof navigator !== 'undefined' && 'clipboard' in navigator;
}

/**
 * Returns a truncated preview of a text string for display in a copy-confirm toast.
 * e.g. 'api-key-very-long-value...' -> 'api-key-very...' (max 20 chars)
 */
export function clipboardPreview(text: string, maxLen: number = 20): string {
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen - 3) + '...';
}

/** Returns a CSS class based on whether clipboard is supported. */
export function clipboardButtonClass(supported: boolean, base: string = 'btn-copy'): string {
  return supported ? base : `${base} ${base}--disabled`;
}
