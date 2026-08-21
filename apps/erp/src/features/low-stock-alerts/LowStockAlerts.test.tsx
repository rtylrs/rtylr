import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import LowStockAlerts from './LowStockAlerts';

const mocks = vi.hoisted(() => ({
  useLowStock: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

describe('LowStockAlerts empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useLowStock.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and navigates to inventory on CTA click', () => {
    render(<LowStockAlerts />);

    const title = screen.getByText('lowStockAlerts.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('lowStockAlerts.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(
      within(emptyState).getByRole('button', { name: 'lowStockAlerts.viewInventory' })
    );

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/inventory' });
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

    expect(screen.queryByText('lowStockAlerts.empty.title')).not.toBeInTheDocument();
  });
});
