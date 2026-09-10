import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Expiry from './Expiry';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useExpiry: vi.fn(),
  refetch: vi.fn(),
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

vi.mock('./useExpiry', () => ({
  useExpiry: mocks.useExpiry,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Expiry empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useExpiry.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the illustration, empty copy, and refetches on CTA click', () => {
    render(<Expiry />);

    const title = screen.getByText('No expiring items');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('No batches with expiry dates tracked')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

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

    expect(screen.queryByText('No expiring items')).not.toBeInTheDocument();
  });
});
