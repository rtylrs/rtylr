import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useShortcutStore } from '../stores/shortcutStore';
import { useCartStore } from '../stores/cartStore';

// Global keyboard shortcuts:
//   cmd/ctrl+k  command palette
//   F1          help
//   F2          discount
//   F3          refund
//   Enter       charge current order (when not inside an input, on an
//               interactive element, or while a modal is open)
//   Escape      close modals / cancel
export function useKeyboardShortcuts() {
  const navigate = useNavigate();
  const toggleCommand = useShortcutStore((s) => s.toggleCommand);
  const setHelp = useShortcutStore((s) => s.setHelp);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement | null)?.tagName?.toLowerCase();
      const isInput =
        tag === 'input' ||
        tag === 'textarea' ||
        (e.target as HTMLElement | null)?.isContentEditable;

      if (e.defaultPrevented) return;
      // cmd/ctrl+k toggles command palette
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        toggleCommand();
        return;
      }

      if (e.key === 'F1') {
        e.preventDefault();
        setHelp(true);
        return;
      }
      if (e.key === 'F2') {
        e.preventDefault();
        // Read cart state live so the listener doesn't need to be
        // re-attached on every cart mutation.
        const { orderId, serverOrderId } = useCartStore.getState();
        navigate(`/discount/${serverOrderId ?? orderId ?? 'current'}`);
        return;
      }
      if (e.key === 'F3') {
        e.preventDefault();
        const { orderId, serverOrderId } = useCartStore.getState();
        navigate(`/refund/${serverOrderId ?? orderId ?? 'current'}`);
        return;
      }

      if (!isInput && e.key === 'Enter' && useCartStore.getState().items.length > 0) {
        // Skip when a button/link has focus (its own click handler fires, so
        // navigating too would double-fire e.g. Hold) or when any modal
        // overlay is present.
        const target = e.target as HTMLElement | null;
        const onInteractive =
          tag === 'button' ||
          tag === 'a' ||
          tag === 'select' ||
          Boolean(target?.closest?.('[role="button"]'));
        const modalOpen = Boolean(document.querySelector('[aria-modal="true"]'));
        if (!onInteractive && !modalOpen) {
          navigate('/payment');
        }
        return;
      }

      if (e.key === 'Escape') {
        setHelp(false);
        useShortcutStore.getState().setCommand(false);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate, toggleCommand, setHelp]);
}
