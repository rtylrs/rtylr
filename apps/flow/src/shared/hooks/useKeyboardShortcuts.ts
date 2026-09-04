import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { useFlowStore } from '../stores/flowStore';

// Keyboard shortcuts: global.
// c = new task, / = search, g i = inbox, g m = my tasks, cmd+k = command bar
export function useKeyboardShortcuts() {
  const navigate = useNavigate();
  // The same gate router.tsx puts on these destinations. Without it the g
  // chords were the remaining way into "Preview route unavailable": the
  // command bar already filters nav-timeline, nav-roadmap and nav-insights,
  // and the sidebar, the project view tabs and Settings all hide them.
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
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
        e.stopPropagation();
        setCommandBarOpen(true);
        return;
      }

      // c => new task
      if (e.key === 'c') {
        if (quickCreateOpen || commandBarOpen) return;
        e.preventDefault();
        // The command bar prints C as the shortcut for New task, so this
        // handler has to be the one that wins. It runs in the capture phase
        // now and stops the event here, because a listener further down that
        // calls stopPropagation would otherwise take the key and the app
        // would not honour the shortcut it advertises.
        e.stopPropagation();
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
          // X6: `g d` retires with the workspace dashboard it pointed at.
          // All projects is reachable from the sidebar; inventing a new chord
          // for it is not what the ledger asked for.
          case 'c':
            navigate({ to: '/calendar' });
            break;
          case 't':
            if (showPreviewRoutes) navigate({ to: '/timeline' });
            break;
          case 'r':
            if (showPreviewRoutes) navigate({ to: '/roadmap' });
            break;
          case 's':
            if (showPreviewRoutes) navigate({ to: '/insights' });
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

    // Capture, not bubble. The typing guard above runs first either way, so
    // a field still receives every key it should.
    window.addEventListener('keydown', handler, true);
    return () => window.removeEventListener('keydown', handler, true);
  }, [
    navigate,
    setCommandBarOpen,
    setQuickCreateOpen,
    commandBarOpen,
    quickCreateOpen,
    activeWorkspaceId,
    showPreviewRoutes,
  ]);
}
