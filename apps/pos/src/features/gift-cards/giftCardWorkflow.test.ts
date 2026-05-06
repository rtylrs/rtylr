import { describe, expect, it, vi } from 'vitest';
import { issueGiftCard } from './giftCardWorkflow';
import type { GiftCard } from '../../shared/types/pos';

const card: GiftCard = {
  id: 'gc_1',
  code: 'GC-1',
  balance: 25,
  issuedAt: '2026-05-06T00:00:00.000Z',
  active: true,
};

describe('issueGiftCard', () => {
  it('returns the issued card when store issuance succeeds', () => {
    expect(
      issueGiftCard({
        codeInput: ' GC-1 ',
        amountInput: '25',
        issue: vi.fn(() => card),
      })
    ).toEqual({ ok: true, card });
  });

  it('returns a validation error when issuance fails', () => {
    expect(
      issueGiftCard({
        codeInput: 'GC-1',
        amountInput: '-10',
        issue: vi.fn(() => {
          throw new Error('Gift card amount must be greater than zero');
        }),
      })
    ).toEqual({ ok: false, error: 'Gift card amount must be greater than zero' });
  });
});
