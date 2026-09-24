import { beforeEach, describe, expect, it, vi } from 'vitest';

const brandClient = vi.hoisted(() => ({ listBrands: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => brandClient,
}));

import { fetchAllBrands } from './useBrands';

describe('fetchAllBrands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses response metadata to fetch brands beyond the first page', async () => {
    brandClient.listBrands
      .mockResolvedValueOnce({
        brands: Array.from({ length: 100 }, (_, index) => ({ id: `b-${index}` })),
        total: 150,
        hasMore: true,
      })
      .mockResolvedValueOnce({
        brands: Array.from({ length: 50 }, (_, index) => ({ id: `b-${100 + index}` })),
        total: 150,
        hasMore: false,
      });

    const brands = await fetchAllBrands();

    expect(brands).toHaveLength(150);
    expect(brandClient.listBrands).toHaveBeenNthCalledWith(1, { limit: 100, offset: 0 });
    expect(brandClient.listBrands).toHaveBeenNthCalledWith(2, { limit: 100, offset: 100 });
  });
});
