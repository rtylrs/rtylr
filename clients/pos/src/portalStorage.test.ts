import { describe, expect, it } from 'vitest';
import { preparePortalTerminal } from './portalStorage';
import { useCartStore } from './shared/stores/cartStore';
import { useOfflineQueueStore } from './shared/stores/offlineQueueStore';
import { useUserStore } from './shared/stores/userStore';

describe('Portal terminal business isolation', () => {
  it('preserves the native cart and restores each business cart and pending sale', async () => {
    useCartStore.getState().addItem({ productId: 10, name: 'Native item', unitPrice: 8 });
    const nativeCart = localStorage.getItem(
      useCartStore.persist.getOptions().name ?? 'pos-active-cart'
    );
    await preparePortalTerminal(101);
    expect(useCartStore.getState().items).toEqual([]);
    useCartStore.getState().addItem({ productId: 11, name: 'Business A item', unitPrice: 9 });
    const sale = useOfflineQueueStore
      .getState()
      .enqueue({ items: [], paymentMethod: 'cash' }, { idempotencyKey: 'sale-business-a' });
    await Promise.all([preparePortalTerminal(101), preparePortalTerminal(101)]);
    expect(useCartStore.getState().items[0].name).toBe('Business A item');
    expect(useOfflineQueueStore.getState().queue[0].id).toBe(sale.id);
    await preparePortalTerminal(102);
    expect(useCartStore.getState().items).toEqual([]);
    expect(useOfflineQueueStore.getState().queue).toEqual([]);
    useCartStore.getState().addItem({ productId: 12, name: 'Business B item', unitPrice: 10 });
    await preparePortalTerminal(101);
    expect(useCartStore.getState().items[0].name).toBe('Business A item');
    expect(useOfflineQueueStore.getState().queue[0].idempotencyKey).toBe('sale-business-a');
    expect(localStorage.getItem('pos-active-cart')).toBe(nativeCart);
    expect(useUserStore.persist.getOptions().name).toBe('portal-pos:101:pos-user-store');
  });
  it('serializes overlapping business changes', async () => {
    await Promise.all([preparePortalTerminal(103), preparePortalTerminal(104)]);
    expect(useCartStore.persist.getOptions().name ?? 'pos-active-cart').toContain(
      'portal-pos:104:'
    );
    expect(useOfflineQueueStore.persist.getOptions().name).toContain('portal-pos:104:');
  });
});
