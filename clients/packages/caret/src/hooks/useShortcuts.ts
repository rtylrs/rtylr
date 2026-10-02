import { useEffect } from 'react';
import { useCaretStore } from '../stores/caretStore';
import { useCaretUIStore } from '../stores/uiStore';
import { requestCaretComposerFocus } from '../lib/focusComposer';

/**
 * Wire global keyboard shortcuts:
 *   Cmd/Ctrl+J  → toggle the Caret panel
 *   Cmd/Ctrl+K  → switch to chat view + focus composer
 *   Esc         → close the panel if it's open and nothing else owns focus
 */
export function useShortcuts(): void {
  const toggle = useCaretStore((s) => s.toggle);
  const close = useCaretStore((s) => s.close);
  const isOpen = useCaretStore((s) => s.isOpen);
  const setView = useCaretUIStore((s) => s.setView);

  useEffect(() => {
    const onKey = (e: KeyboardEvent): void => {
      if (e.defaultPrevented) return;
      const mod = e.metaKey || e.ctrlKey;
      const target = e.target as HTMLElement | null;
      const editable =
        !!target &&
        (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable);

      if (mod && (e.key === 'j' || e.key === 'J')) {
        e.preventDefault();
        toggle();
        return;
      }

      if (mod && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault();
        setView('chat');
        if (!isOpen) toggle();
        requestCaretComposerFocus();
        return;
      }

      if (e.key === 'Escape' && isOpen && !editable) {
        close();
      }
    };
    window.addEventListener('keydown', onKey);
    return (): void => window.removeEventListener('keydown', onKey);
  }, [toggle, close, isOpen, setView]);
}
