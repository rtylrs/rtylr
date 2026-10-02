import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fetchAllSuppliers } from './useSuppliers';

const mocks = vi.hoisted(() => ({ listSuppliers: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ listSuppliers: mocks.listSuppliers }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn() } }));

describe('fetchAllSuppliers', () => {
  beforeEach(() => mocks.listSuppliers.mockReset());

  it('follows hasMore until every supplier page is loaded', async () => {
    mocks.listSuppliers
      .mockResolvedValueOnce({
        suppliers: Array.from({ length: 50 }, (_, index) => ({
          id: String(index + 1),
          name: `S${index + 1}`,
        })),
        hasMore: true,
      })
      .mockResolvedValueOnce({
        suppliers: [{ id: '51', name: 'S51' }],
        hasMore: false,
      });

    const response = await fetchAllSuppliers(50);
    expect(response.suppliers).toHaveLength(51);
    expect(mocks.listSuppliers).toHaveBeenNthCalledWith(1, { page: 1, pageSize: 50 });
    expect(mocks.listSuppliers).toHaveBeenNthCalledWith(2, { page: 2, pageSize: 50 });
  });

  it('rejects a repeated page rather than presenting a truncated directory', async () => {
    mocks.listSuppliers.mockResolvedValue({ suppliers: [{ id: '1', name: 'One' }], hasMore: true });
    await expect(fetchAllSuppliers(50)).rejects.toThrow('Supplier pagination did not advance');
  });
});
