import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Expiry from './Expiry';

const mocks = vi.hoisted(() => ({
  useExpiry: vi.fn(),
  refetch: vi.fn(),
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

vi.mock('./useExpiry', () => ({
  useExpiry: mocks.useExpiry,
}));

describe('Expiry empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useExpiry.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the illustration, empty copy, and refetches on CTA click', () => {
    render(<Expiry />);

    const title = screen.getByText('expiry.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('expiry.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'expiry.refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when items exist', () => {
    mocks.useExpiry.mockReturnValue({
      data: [
        {
          productId: 'p1',
          batchId: 'b1',
          name: 'Milk',
          sku: 'M-1',
          quantity: 5,
          warehouseName: 'Main',
          expiryDate: '2026-09-01',
        },
      ],
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<Expiry />);

    expect(screen.queryByText('expiry.empty.title')).not.toBeInTheDocument();
  });
});
