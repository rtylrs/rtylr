import { describe, expect, it } from 'vitest';
import { normalizeRemoteGiftCard } from './giftCardRemote';

describe('normalizeRemoteGiftCard', () => {
  it('normalizes direct and nested gift card response shapes', () => {
    expect(
      normalizeRemoteGiftCard({
        card: {
          id: 7,
          code: ' GC-7 ',
          balance: '25.50',
          issuedAt: '2026-01-01T00:00:00.000Z',
          active: true,
          customerName: 'Ada',
        },
      }),
    ).toEqual({
      id: '7',
      code: 'GC-7',
      balance: 25.5,
      issuedAt: '2026-01-01T00:00:00.000Z',
      active: true,
      customerName: 'Ada',
    });

    expect(
      normalizeRemoteGiftCard({
        data: {
          id: 'gc_8',
          code: 'GC-8',
          balance: 10,
          issued_at: '2026-01-02T00:00:00.000Z',
          is_active: false,
        },
      }),
    ).toMatchObject({
      id: 'gc_8',
      code: 'GC-8',
      balance: 10,
      issuedAt: '2026-01-02T00:00:00.000Z',
      active: false,
    });
  });

  it('returns null for unusable remote gift card data', () => {
    expect(normalizeRemoteGiftCard({ data: { code: 'GC', balance: 'bad' } })).toBeNull();
    expect(normalizeRemoteGiftCard(null)).toBeNull();
  });
});
