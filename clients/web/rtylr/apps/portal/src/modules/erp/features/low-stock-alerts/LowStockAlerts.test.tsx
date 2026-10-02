import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import LowStockAlerts from './LowStockAlerts';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useLowStock: vi.fn(),
  navigate: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('../inventory/useInventory', () => ({
  useLowStock: mocks.useLowStock,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('LowStockAlerts empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useLowStock.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and navigates to inventory on CTA click', () => {
    render(<LowStockAlerts />);

    const title = screen.getByText('All stock levels healthy');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Nothing needs reordering right now.')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'View Inventory' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/erp/inventory' });
  });

  it('does not show the empty state when there are low-stock items', () => {
    mocks.useLowStock.mockReturnValue({
      data: [
        {
          productId: 'p1',
          productName: 'Widget',
          sku: 'W-1',
          stockQuantity: 2,
          lowStockThreshold: 10,
          unit: 'ea',
        },
      ],
      isLoading: false,
    });

    render(<LowStockAlerts />);

    expect(screen.queryByText('All stock levels healthy')).not.toBeInTheDocument();
  });
});
