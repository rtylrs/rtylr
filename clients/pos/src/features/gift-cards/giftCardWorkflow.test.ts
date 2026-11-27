import { describe, expect, it, vi } from 'vitest';

const giftCardClient = vi.hoisted(() => ({
  createGiftCard: vi.fn(() => Promise.resolve({ giftCardId: 'remote-gc', message: 'ok' })),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => giftCardClient,
}));

import { generateGiftCardCode, issueGiftCard } from './giftCardWorkflow';
import type { GiftCard } from '../../shared/types/pos';

const card: GiftCard = {
  id: 'gc_1',
  code: 'GC-1',
  balance: 25,
  issuedAt: '2026-05-06T00:00:00.000Z',
  active: true,
};

describe('issueGiftCard', () => {
  it('returns the issued card when store issuance succeeds', async () => {
    await expect(
      issueGiftCard({
        codeInput: ' GC-1 ',
        amountInput: '25',
        issue: vi.fn(() => card),
      })
    ).resolves.toEqual({ ok: true, card });
  });

  it('returns a validation error for negative amounts', async () => {
    await expect(
      issueGiftCard({
        codeInput: 'GC-1',
        amountInput: '-10',
        issue: vi.fn(),
      })
    ).resolves.toEqual({ ok: false, error: 'Gift card amount must be greater than zero' });
  });

  it('returns a validation error for empty code', async () => {
    await expect(
      issueGiftCard({
        codeInput: '  ',
        amountInput: '25',
        issue: vi.fn(),
      })
    ).resolves.toEqual({ ok: false, error: 'Code and amount required' });
  });

  it('captures store issuance errors and returns ok: false', async () => {
    await expect(
      issueGiftCard({
        codeInput: 'GC-1',
        amountInput: '25',
        issue: vi.fn(() => {
          throw new Error('Gift card amount must be greater than zero');
        }),
      })
    ).resolves.toEqual({ ok: false, error: 'Gift card amount must be greater than zero' });
  });

  it('sends YYYY-MM-DD expiresAt, real paidOrderId, and customer details to the API', async () => {
    giftCardClient.createGiftCard.mockClear();
    await issueGiftCard({
      codeInput: 'GC-REAL-1',
      amountInput: '40',
      paidOrderId: '99',
      customerName: 'Maya',
      customerEmail: 'maya@example.com',
      expiresAt: '2027-06-15T12:00:00Z',
      issue: vi.fn(() => card),
    });

    expect(giftCardClient.createGiftCard).toHaveBeenCalledWith({
      code: 'GC-REAL-1',
      paidOrderId: '99',
      initialAmount: 40,
      customerName: 'Maya',
      customerEmail: 'maya@example.com',
      expiresAt: '2027-06-15',
    });
  });

  it('generates formatted gift card codes', () => {
    expect(generateGiftCardCode()).toMatch(/^GC-[A-Z0-9]{4}-[A-Z0-9]{4}$/);
  });
});
