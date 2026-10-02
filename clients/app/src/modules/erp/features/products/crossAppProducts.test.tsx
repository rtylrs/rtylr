import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import { useMenuItems } from '../../../../../../pos/src/features/menu-items/useMenuItems';
import { useCreateProduct } from './useProducts';

const api = vi.hoisted(() => ({
  createProduct: vi.fn(),
  listPOSCatalog: vi.fn(),
}));

vi.mock('@voxire/api', () => ({ createOneKitClient: () => api }));

describe('ERP mutations refresh the POS catalog', () => {
  it('shows a newly created ERP product while the POS catalog cache is still fresh', async () => {
    const items = [{ id: '1', name: 'Existing product', salePrice: 5 }];
    api.listPOSCatalog.mockImplementation(async () => ({ items: [...items] }));
    api.createProduct.mockImplementation(async () => {
      items.push({ id: '2', name: 'New ERP product', salePrice: 8 });
      return { id: '2' };
    });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: React.PropsWithChildren) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const { result, unmount } = renderHook(
      () => ({ catalog: useMenuItems(), create: useCreateProduct() }),
      { wrapper }
    );
    await waitFor(() => expect(result.current.catalog.items).toHaveLength(1));
    await act(async () => {
      await result.current.create.mutateAsync({
        name: 'New ERP product',
        salePrice: 8,
        categoryId: '1',
        supplierId: '1',
        sku: 'CROSS-2',
        barcode: 'CROSS-2',
        description: 'Cross-app fixture',
        unit: 'pcs',
      });
    });
    await waitFor(() =>
      expect(result.current.catalog.items.map((item) => item.name)).toEqual([
        'Existing product',
        'New ERP product',
      ])
    );
    unmount();
    queryClient.clear();
  });
});
