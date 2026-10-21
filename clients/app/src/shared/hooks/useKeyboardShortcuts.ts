import { useEffect, useRef } from 'react';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useCommandPalette } from '../stores/commandPaletteStore';
import { useUserStore } from '../stores/userStore';
import { canAccessPortalPath } from '../utils/canAccessPortalPath';

export const useKeyboardShortcuts = () => {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const { toggle, setOpen } = useCommandPalette();
  const chordRef = useRef<{ key: string; at: number } | null>(null);

  useEffect(() => {
    const isEditable = (el: Element | null): boolean => {
      if (!el) return false;
      const tag = el.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true;
      return (el as HTMLElement).isContentEditable;
    };

    const handler = (e: KeyboardEvent) => {
      if (e.defaultPrevented || e.repeat) return;
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        chordRef.current = null;
        toggle();
        return;
      }

      if (useCommandPalette.getState().open || isEditable(document.activeElement)) return;
      const go = (path: string) => {
        if (!canAccessPortalPath(useUserStore.getState().user, path)) return;
        e.preventDefault();
        navigate({ to: path });
      };

      if (e.key === '?') {
        e.preventDefault();
        setOpen(true);
        return;
      }

      if (e.key.toLowerCase() === 'n' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        if (pathname.startsWith('/hr')) go('/hr/employees/new');
        else if (pathname.startsWith('/finance')) go('/finance/invoices/new');
        else if (!/^\/(flow|pos|crm|erp|recruit|insights|caret)(\/|$)/.test(pathname))
          go('/team/invite');
        return;
      }

      if (!e.metaKey && !e.ctrlKey && !e.altKey && !chordRef.current) {
        if (pathname.startsWith('/hr')) {
          if (e.key === '/') {
            e.preventDefault();
            window.dispatchEvent(new CustomEvent('hr:focus-search'));
            return;
          }
          if (e.key.toLowerCase() === 't') {
            go('/hr/time-off');
            return;
          }
          if (e.key.toLowerCase() === 's') {
            go('/hr/schedule');
            return;
          }
        }
        if (pathname.startsWith('/finance') && e.key.toLowerCase() === 'e') {
          go('/finance/expenses/new');
          return;
        }
      }

      if (e.key.toLowerCase() === 'g' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        chordRef.current = { key: 'g', at: Date.now() };
        return;
      }

      if (chordRef.current && chordRef.current.key === 'g') {
        const elapsed = Date.now() - chordRef.current.at;
        if (elapsed > 1500) {
          chordRef.current = null;
          return;
        }
        const k = e.key.toLowerCase();
        const routes: Record<string, string> = {
          d: '/',
          t: '/team',
          s: '/settings/account',
          n: '/notifications',
          b: '/settings/billing',
          i: '/extensions',
          k: '/api-keys',
          h: '/audit',
          u: '/audit',
          r: '/roles',
        };
        if (routes[k]) {
          go(routes[k]);
        }
        chordRef.current = null;
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [navigate, pathname, toggle, setOpen]);
};
