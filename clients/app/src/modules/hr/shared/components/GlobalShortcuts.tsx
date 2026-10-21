import React, { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';

/**
 * Registers keyboard shortcuts across the app.
 * n = new hire, / = focus search (dispatches an event), t = time off, s = shifts
 */
const GlobalShortcuts: React.FC = () => {
  const navigate = useNavigate();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Ignore shortcuts while typing in inputs
      const target = e.target as HTMLElement;
      const tag = target.tagName?.toLowerCase();
      if (tag === 'input' || tag === 'textarea' || target.isContentEditable) return;
      // Ignore if modifier is held (cmd+K handled separately)
      if (e.metaKey || e.ctrlKey || e.altKey) return;

      switch (e.key.toLowerCase()) {
        case 'n':
          navigate({ to: '/hr/staff' });
          toast('Tip: click "New hire" to add a staff member', { duration: 2000 });
          break;
        case '/':
          e.preventDefault();
          window.dispatchEvent(new CustomEvent('hr:focus-search'));
          break;
        case 't':
          navigate({ to: '/hr/time-off' });
          break;
        case 's':
          navigate({ to: '/hr/schedule' });
          break;
        default:
          break;
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate]);

  return null;
};

export default GlobalShortcuts;
