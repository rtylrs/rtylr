/**
 * Centralized list of Caret-related keyboard shortcuts so the help dialog
 * and the shortcut handler stay in sync.
 */

export interface ShortcutDef {
  id: string;
  keys: string[];
  description: string;
}

export const CARET_SHORTCUTS: ShortcutDef[] = [
  { id: 'toggle', keys: ['⌘', 'J'], description: 'Toggle Caret panel' },
  { id: 'focus', keys: ['⌘', 'K'], description: 'Open Caret and focus the composer' },
  { id: 'newline', keys: ['Shift', 'Enter'], description: 'New line without sending' },
  { id: 'send', keys: ['Enter'], description: 'Send message' },
  { id: 'close', keys: ['Esc'], description: 'Close Caret (when not typing)' },
];
