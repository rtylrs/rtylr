import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { validateAdjustInventoryRequest } from '@voxire/voxie/erp/inventory/v1/types';
import QuickStockAdjust from './QuickStockAdjust';
import { initTestI18n } from '../testing/i18n';

const mocks = vi.hoisted(() => ({
  adjust: vi.fn(),
}));

vi.mock('../../features/inventory/useInventory', () => ({
  useAdjustInventory: () => ({ mutate: mocks.adjust, isPending: false }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('QuickStockAdjust', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not submit a decrease the API contract rejects', () => {
    render(<QuickStockAdjust productId="42" currentStock={4} />);

    const decrease = screen.getByRole('button', { name: 'Decrease stock' });
    expect(decrease).toBeDisabled();
    expect(decrease).toHaveAttribute('title');

    fireEvent.click(decrease);
    expect(mocks.adjust).not.toHaveBeenCalled();
  });

  it('submits a positive change that passes request validation for an increase', async () => {
    const user = userEvent.setup();
    render(<QuickStockAdjust productId="42" currentStock={4} />);

    await user.click(screen.getByRole('button', { name: 'Increase stock' }));

    expect(mocks.adjust).toHaveBeenCalledTimes(1);
    const [request] = mocks.adjust.mock.calls[0];
    expect(validateAdjustInventoryRequest(request)).toEqual([]);
    expect(request).toEqual({
      productId: '42',
      change: 1,
      type: 'adjustment',
      notes: 'Quick stock adjustment',
    });
    expect(screen.getByText('5')).toBeInTheDocument();
  });
});
