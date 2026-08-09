/** Returns true if the modifier key (Ctrl or Cmd) is held. */
export function isModKey(e: KeyboardEvent): boolean {
  return e.ctrlKey || e.metaKey;
}

/** Returns true if the keyboard event is for the given key (case-insensitive). */
export function isKey(e: KeyboardEvent, key: string): boolean {
  return e.key.toLowerCase() === key.toLowerCase();
}

/** Returns true if the event is a confirm action (Enter without shift). */
export function isConfirmKey(e: KeyboardEvent): boolean {
  return e.key === 'Enter' && !e.shiftKey;
}

/** Returns true if the event is an Escape key press. */
export function isEscapeKey(e: KeyboardEvent): boolean {
  return e.key === 'Escape';
}

/** Returns true if the event is an arrow navigation key. */
export function isArrowKey(e: KeyboardEvent): boolean {
  return ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(e.key);
}

/** Returns true if the event is for selecting all (Ctrl/Cmd + A). */
export function isSelectAll(e: KeyboardEvent): boolean {
  return isModKey(e) && isKey(e, 'a');
}

/** Returns true if the event targets an input or textarea element. */
export function isInInputField(e: KeyboardEvent): boolean {
  const target = e.target as HTMLElement;
  return target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;
}

/** A mapping of common keyboard shortcut descriptions. */
export const KEYBOARD_SHORTCUTS: Record<string, string> = {
  c: 'Create task',
  n: 'New project',
  '/': 'Search',
  f: 'Filter',
  'g i': 'Go to Inbox',
  'g m': 'Go to My Tasks',
  Escape: 'Close dialog',
} as const;

/** Returns the shortcut description for a key, or null. */
export function shortcutDescription(key: string): string | null {
  return KEYBOARD_SHORTCUTS[key] ?? null;
}
