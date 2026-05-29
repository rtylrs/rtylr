/**
 * Keyboard shortcut utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type ModifierKey = 'ctrl' | 'alt' | 'shift' | 'meta';

export interface KeyCombo {
  key: string;
  modifiers?: ModifierKey[];
}

/** Return a human-readable string for a key combo (e.g., "Ctrl+K"). */
export function formatKeyCombo(combo: KeyCombo): string {
  const mods = (combo.modifiers ?? []).map((m) => {
    switch (m) {
      case 'ctrl':
        return 'Ctrl';
      case 'alt':
        return 'Alt';
      case 'shift':
        return 'Shift';
      case 'meta':
        return 'Cmd';
      default:
        return m;
    }
  });
  return [...mods, combo.key.toUpperCase()].join('+');
}

/** Return true when a keyboard event matches a key combo. */
export function matchesCombo(event: KeyboardEvent, combo: KeyCombo): boolean {
  const mods = combo.modifiers ?? [];
  const ctrlOk = mods.includes('ctrl') ? event.ctrlKey : !event.ctrlKey;
  const altOk = mods.includes('alt') ? event.altKey : !event.altKey;
  const shiftOk = mods.includes('shift') ? event.shiftKey : !event.shiftKey;
  const metaOk = mods.includes('meta') ? event.metaKey : !event.metaKey;
  const keyOk = event.key.toLowerCase() === combo.key.toLowerCase();
  return ctrlOk && altOk && shiftOk && metaOk && keyOk;
}

/** Return true when the key is a navigation key (arrows, home, end, etc.). */
export function isNavigationKey(key: string): boolean {
  return [
    'ArrowUp',
    'ArrowDown',
    'ArrowLeft',
    'ArrowRight',
    'Home',
    'End',
    'PageUp',
    'PageDown',
  ].includes(key);
}

/** Return true when the key triggers a form submission or action. */
export function isActionKey(key: string): boolean {
  return key === 'Enter' || key === ' ';
}

/** Return true when the key dismisses a UI element. */
export function isDismissKey(key: string): boolean {
  return key === 'Escape';
}

/** Return true when the key moves focus to the next element. */
export function isForwardTab(event: KeyboardEvent): boolean {
  return event.key === 'Tab' && !event.shiftKey;
}

/** Return true when the key moves focus to the previous element. */
export function isBackwardTab(event: KeyboardEvent): boolean {
  return event.key === 'Tab' && event.shiftKey;
}
