import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useFlowStore } from '../stores/flowStore';

// Keyboard shortcuts: global.
// c = new task, / = search, g i = inbox, g m = my tasks, cmd+k = command bar
export function useKeyboardShortcuts() {
  const navigate = useNavigate();
  const {
    setCommandBarOpen,
    setQuickCreateOpen,
    commandBarOpen,
    quickCreateOpen,
    activeWorkspaceId,
  } = useFlowStore();

  useEffect(() => {
    let lastKey = '';
    let lastKeyAt = 0;

    const isTypingContext = (target: EventTarget | null) => {
      if (!(target instanceof HTMLElement)) return false;
      const tag = target.tagName;
      return (
        tag === 'INPUT' ||
        tag === 'TEXTAREA' ||
        target.isContentEditable ||
        target.getAttribute('role') === 'textbox'
      );
    };

    const handler = (e: KeyboardEvent) => {
      // cmd/ctrl + k always opens command bar
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setCommandBarOpen(!commandBarOpen);
        return;
      }

      // ignore typing contexts for single-letter shortcuts
      if (isTypingContext(e.target) || e.metaKey || e.ctrlKey || e.altKey) return;

      // /
      if (e.key === '/') {
        e.preventDefault();
        setCommandBarOpen(true);
        return;
      }

      // c => new task
      if (e.key === 'c') {
        if (quickCreateOpen || commandBarOpen) return;
        e.preventDefault();
        setQuickCreateOpen(true);
        return;
      }

      // g + X chord navigation
      const now = Date.now();
      if (lastKey === 'g' && now - lastKeyAt < 1200) {
        e.preventDefault();
        switch (e.key) {
          case 'i':
            navigate({ to: '/inbox' });
            break;
          case 'm':
            navigate({ to: '/my-issues' });
            break;
          case 'd':
            // Was the workspace dashboard, which never had a route of its own.
            // All projects is the reachability replacement for it.
            navigate({ to: '/projects' });
            break;
          case 'c':
            navigate({ to: '/calendar' });
            break;
          case 't':
            navigate({ to: '/timeline' });
            break;
          case 'r':
            navigate({ to: '/roadmap' });
            break;
          case 's':
            navigate({ to: '/insights' });
            break;
          case 'y':
            navigate({ to: '/cycles' });
            break;
          case 'x':
            navigate({ to: '/triage' });
            break;
          case 'f':
            navigate({ to: '/today' });
            break;
        }
        lastKey = '';
        return;
      }

      if (e.key === 'g') {
        lastKey = 'g';
        lastKeyAt = now;
      } else {
        lastKey = e.key;
        lastKeyAt = now;
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [
    navigate,
    setCommandBarOpen,
    setQuickCreateOpen,
    commandBarOpen,
    quickCreateOpen,
    activeWorkspaceId,
  ]);
}
