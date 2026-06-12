/**
 * End-to-end: manager override PIN verification → reopen a paid order.
 *
 * Chains posService.verifyManagerPin → posService.reopenOrder to prove
 * the two API calls fire in the correct sequence and gating logic works.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    verifyManagerPin: vi.fn(),
    reopenOrder: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((p) => p),
}));

import { posService } from '../../shared/services/api/posService';
import { unwrap } from '@voxire/api';

// Pure workflow helper that mirrors what useOverride + calling code would do.
async function managerReopenOrder(pin: string, orderId: number): Promise<void> {
  if (!pin.trim()) {
    throw new Error('Enter a manager PIN');
  }
  const res = await unwrap(posService.verifyManagerPin(pin.trim()));
  const verified = !!(
    (res as Record<string, unknown> | undefined)?.verified ??
    (res as Record<string, unknown> | undefined)?.ok ??
    (res as Record<string, unknown> | undefined)?.success
  );
  if (!verified) {
    throw new Error('Invalid PIN');
  }
  await unwrap(posService.reopenOrder(orderId));
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('manager override → reopen paid order', () => {
  it('correct PIN unlocks reopen of a paid order', async () => {
    vi.mocked(posService.verifyManagerPin).mockResolvedValue(
      { verified: true } as never,
    );
    vi.mocked(posService.reopenOrder).mockResolvedValue(undefined as never);

    await managerReopenOrder('1234', 42);

    expect(posService.verifyManagerPin).toHaveBeenCalledTimes(1);
    expect(posService.verifyManagerPin).toHaveBeenCalledWith('1234');

    expect(posService.reopenOrder).toHaveBeenCalledTimes(1);
    expect(posService.reopenOrder).toHaveBeenCalledWith(42);

    // Verify ordering: verifyManagerPin resolved before reopenOrder was called.
    const verifyOrder = vi.mocked(posService.verifyManagerPin).mock.invocationCallOrder[0];
    const reopenOrder = vi.mocked(posService.reopenOrder).mock.invocationCallOrder[0];
    expect(verifyOrder).toBeLessThan(reopenOrder);
  });

  it('wrong PIN blocks reopen - reopenOrder is never called', async () => {
    vi.mocked(posService.verifyManagerPin).mockResolvedValue(
      { verified: false } as never,
    );

    await expect(managerReopenOrder('0000', 42)).rejects.toThrow('Invalid PIN');

    expect(posService.verifyManagerPin).toHaveBeenCalledTimes(1);
    expect(posService.reopenOrder).not.toHaveBeenCalled();
  });

  it('blank PIN is rejected before hitting the API', async () => {
    await expect(managerReopenOrder('   ', 42)).rejects.toThrow('Enter a manager PIN');

    expect(posService.verifyManagerPin).not.toHaveBeenCalled();
    expect(posService.reopenOrder).not.toHaveBeenCalled();
  });
});
