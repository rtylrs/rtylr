import React from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Barcodes from './Barcodes';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
  fetchAllProducts: vi.fn(),
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

vi.mock('../products/useProducts', () => ({
  useProducts: mocks.useProducts,
  fetchAllProducts: mocks.fetchAllProducts,
}));

vi.mock('../../shared/components/BarcodeSvg', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

function renderBarcodes() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <Barcodes />
    </QueryClientProvider>
  );
}

describe('Barcodes preview empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useProducts.mockReturnValue({
      data: { products: [] },
      isLoading: false,
    });
    mocks.fetchAllProducts.mockResolvedValue([]);
  });

  it('selects the first product on CTA click when products are available', async () => {
    mocks.fetchAllProducts.mockResolvedValue([
      { id: 'p1', name: 'Widget', sku: 'W-1', salePrice: 10 },
    ]);

    renderBarcodes();

    await screen.findByText('Widget');
    const title = screen.getByText('Select products to preview labels');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText(
        'Choose one or more products from the list to see their barcode labels here.'
      )
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Select First Product' }));

    expect(screen.queryByText('Select products to preview labels')).not.toBeInTheDocument();
  });

  it('loads products beyond the first server page for the picker', async () => {
    mocks.useProducts.mockReturnValue({
      data: { products: [{ id: 'p1', name: 'Product 1', sku: 'P-1', salePrice: 10 }] },
      isLoading: false,
    });
    mocks.fetchAllProducts.mockResolvedValue([
      { id: 'p1', name: 'Product 1', sku: 'P-1', salePrice: 10 },
      { id: 'p120', name: 'Product 120', sku: 'P-120', salePrice: 12 },
    ]);

    renderBarcodes();

    expect(await screen.findByText('Product 120')).toBeInTheDocument();
    expect(mocks.fetchAllProducts).toHaveBeenCalled();
  });

  it('navigates to products on CTA click when there are no products at all', async () => {
    renderBarcodes();

    await waitFor(() => expect(mocks.fetchAllProducts).toHaveBeenCalled());
    const title = screen.getByText('Select products to preview labels');
    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Go to Products' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/erp/products' });
  });
});
