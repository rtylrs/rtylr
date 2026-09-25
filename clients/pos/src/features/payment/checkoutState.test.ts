import { describe, expect, it, vi } from 'vitest';
import { resetAfterCheckout } from './checkoutState';

describe('resetAfterCheckout', () => {
  it('starts a new order instead of clearing terminal context', () => {
    const newOrder = vi.fn();
    const clearCart = vi.fn();

    resetAfterCheckout({ newOrder, clearCart });

    expect(newOrder).toHaveBeenCalledOnce();
    expect(clearCart).not.toHaveBeenCalled();
  });
});
