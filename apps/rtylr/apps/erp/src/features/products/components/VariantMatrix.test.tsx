import React from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import VariantMatrix from './VariantMatrix';
import { initTestI18n } from '../../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useVariants: vi.fn(),
  createVariant: { mutateAsync: vi.fn(), isPending: false },
  toast: { success: vi.fn(), error: vi.fn() },
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('../useVariants', () => ({
  parseVariantAttributes: (raw: string) => {
    try {
      return raw ? (JSON.parse(raw) as Record<string, string>) : {};
    } catch {
      return {};
    }
  },
  useVariants: mocks.useVariants,
  useCreateVariant: () => mocks.createVariant,
  useUpdateVariant: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteVariant: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('sonner', () => ({ toast: mocks.toast }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('VariantMatrix empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [] });
  });

  it('shows the illustration and empty copy, and adds an option group on CTA click', () => {
    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    const title = screen.getByText('No variants yet. Add an option below to generate them.');
    expect(title).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Add option' }));

    expect(screen.getByPlaceholderText('Option name (e.g. Size, Color)')).toBeInTheDocument();
    expect(
      screen.queryByText('No variants yet. Add an option below to generate them.')
    ).not.toBeInTheDocument();
  });

  it('does not show the empty state when variants exist', () => {
    mocks.useVariants.mockReturnValue({
      data: [{ id: 'v1', name: 'Large', sku: 'V-1', stockQuantity: 5, salePrice: 10 }],
    });

    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    expect(
      screen.queryByText('No variants yet. Add an option below to generate them.')
    ).not.toBeInTheDocument();
  });
});

describe('VariantMatrix create all', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [] });
    mocks.createVariant.mutateAsync.mockReset();
  });

  it('keeps only uncreated drafts after a failure so a retry cannot duplicate', async () => {
    const user = userEvent.setup();
    mocks.createVariant.mutateAsync
      .mockResolvedValueOnce({})
      .mockRejectedValueOnce(new Error('create failed'))
      .mockResolvedValueOnce({});

    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    const [addOption] = screen.getAllByRole('button', { name: 'Add option' });
    fireEvent.click(addOption);
    await user.type(screen.getByPlaceholderText('Option name (e.g. Size, Color)'), 'Size');
    await user.type(screen.getByPlaceholderText('Add value, press Enter'), 'S{Enter}M{Enter}');
    await user.click(screen.getByRole('button', { name: 'Generate 2 new' }));
    await user.click(screen.getByRole('button', { name: 'Create 2 variants' }));

    await waitFor(() =>
      expect(mocks.toast.error).toHaveBeenCalledWith(
        'Created 1 of 2 variants; retry the remaining rows'
      )
    );
    expect(screen.getByRole('button', { name: 'Create 1 variants' })).toBeInTheDocument();
    expect(screen.getByLabelText('SKU')).toHaveValue('SKU-1-M');

    await user.click(screen.getByRole('button', { name: 'Create 1 variants' }));

    await waitFor(() => expect(mocks.toast.success).toHaveBeenCalledWith('1 variants created'));
    expect(mocks.createVariant.mutateAsync.mock.calls.map(([request]) => request.sku)).toEqual([
      'SKU-1-S',
      'SKU-1-M',
      'SKU-1-M',
    ]);
  });
});
