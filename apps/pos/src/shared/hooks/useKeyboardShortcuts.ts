import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useShortcutStore } from '../stores/shortcutStore';
import { useCartStore } from '../stores/cartStore';

// Global keyboard shortcuts:
//   cmd/ctrl+k  command palette
//   F1          help
//   F2          discount
//   F3          refund
//   Enter       charge current order (when not inside an input)
//   Escape      close modals / cancel
export function useKeyboardShortcuts() {
  const navigate = useNavigate();
  const toggleCommand = useShortcutStore((s) => s.toggleCommand);
  const setHelp = useShortcutStore((s) => s.setHelp);
  const items = useCartStore((s) => s.items);
  const orderId = useCartStore((s) => s.orderId);
  const serverOrderId = useCartStore((s) => s.serverOrderId);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement | null)?.tagName?.toLowerCase();
      const isInput =
        tag === 'input' ||
        tag === 'textarea' ||
        (e.target as HTMLElement | null)?.isContentEditable;

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
        navigate(`/discount/${serverOrderId ?? orderId ?? 'current'}`);
        return;
      }
      if (e.key === 'F3') {
        e.preventDefault();
        navigate(`/refund/${serverOrderId ?? orderId ?? 'current'}`);
        return;
      }

      if (!isInput && e.key === 'Enter' && items.length > 0) {
        navigate('/payment');
        return;
      }

      if (e.key === 'Escape') {
        setHelp(false);
        useShortcutStore.getState().setCommand(false);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate, toggleCommand, setHelp, items.length, orderId, serverOrderId]);
}
