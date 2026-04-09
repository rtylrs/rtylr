import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFlowStore } from '../stores/flowStore';

// Linear-style keyboard shortcuts: global.
// c = new task, / = search, g i = goto inbox, g m = my issues, cmd+k = command bar
export function useKeyboardShortcuts() {
  const navigate = useNavigate();
  const { setCommandBarOpen, setQuickCreateOpen, commandBarOpen, quickCreateOpen, activeWorkspaceId } =
    useFlowStore();

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
        switch (e.key) {
          case 'i':
            navigate('/inbox');
            break;
          case 'm':
            navigate('/my-issues');
            break;
          case 'd':
            navigate(`/w/${activeWorkspaceId}`);
            break;
          case 'c':
            navigate('/calendar');
            break;
          case 't':
            navigate('/timeline');
            break;
          case 'r':
            navigate('/roadmap');
            break;
          case 's':
            navigate('/insights');
            break;
          case 'y':
            navigate('/cycles');
            break;
          case 'x':
            navigate('/triage');
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
  }, [navigate, setCommandBarOpen, setQuickCreateOpen, commandBarOpen, quickCreateOpen, activeWorkspaceId]);
}
