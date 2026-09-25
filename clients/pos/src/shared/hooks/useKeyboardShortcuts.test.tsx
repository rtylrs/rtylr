import { cleanup, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { useCartStore } from '../stores/cartStore';
import { useShortcutStore } from '../stores/shortcutStore';
import type { CartItem } from '../stores/cartStore';

const navigate = vi.hoisted(() => vi.fn());

vi.mock('react-router-dom', () => ({
  useNavigate: () => navigate,
}));

const item = (overrides: Partial<CartItem> = {}): CartItem => ({
  id: '1::',
  lineId: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 10,
  price: 10,
  quantity: 1,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  ...overrides,
});

describe('useKeyboardShortcuts', () => {
  beforeEach(() => {
    navigate.mockClear();
    useShortcutStore.setState({ commandOpen: false, helpOpen: false });
    useCartStore.setState({ items: [], orderId: null, serverOrderId: null });
  });

  afterEach(() => {
    cleanup();
    document.body.innerHTML = '';
  });

  it('toggles command palette with ctrl+k', () => {
    renderHook(() => useKeyboardShortcuts());

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }));

    expect(useShortcutStore.getState().commandOpen).toBe(true);
  });

  it('charges with Enter only when a cart has items and focus is not in an input', () => {
    useCartStore.setState({ items: [item()] });
    renderHook(() => useKeyboardShortcuts());

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));

    expect(navigate).toHaveBeenCalledWith('/payment');

    navigate.mockClear();
    const input = document.createElement('input');
    document.body.append(input);
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(navigate).not.toHaveBeenCalled();
  });

  it('does not charge with Enter while a focused button would handle the key', () => {
    useCartStore.setState({ items: [item()] });
    renderHook(() => useKeyboardShortcuts());

    const button = document.createElement('button');
    document.body.append(button);
    button.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(navigate).not.toHaveBeenCalled();
  });

  it('does not charge with Enter while a modal overlay is open', () => {
    useCartStore.setState({ items: [item()] });
    renderHook(() => useKeyboardShortcuts());

    const overlay = document.createElement('div');
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    document.body.append(overlay);
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));

    expect(navigate).not.toHaveBeenCalled();

    overlay.remove();
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
    expect(navigate).toHaveBeenCalledWith('/payment');
  });

  it('routes discount and refund shortcuts to the synced server order when available', () => {
    useCartStore.setState({ serverOrderId: 42, orderId: null });
    renderHook(() => useKeyboardShortcuts());

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'F2' }));
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'F3' }));

    expect(navigate).toHaveBeenNthCalledWith(1, '/discount/42');
    expect(navigate).toHaveBeenNthCalledWith(2, '/refund/42');
  });

  it('closes shortcut overlays with Escape', () => {
    useShortcutStore.setState({ commandOpen: true, helpOpen: true });
    renderHook(() => useKeyboardShortcuts());

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));

    expect(useShortcutStore.getState().commandOpen).toBe(false);
    expect(useShortcutStore.getState().helpOpen).toBe(false);
  });
});
