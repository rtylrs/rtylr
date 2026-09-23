import { beforeEach, describe, expect, it, vi } from 'vitest';

const categoryClient = vi.hoisted(() => ({ listCategories: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => categoryClient,
}));

import { fetchAllCategories } from './useCategories';

describe('fetchAllCategories', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses response metadata to fetch categories beyond the first page', async () => {
    categoryClient.listCategories
      .mockResolvedValueOnce({
        categories: Array.from({ length: 100 }, (_, index) => ({ id: `c-${index}` })),
        total: 150,
        hasMore: true,
      })
      .mockResolvedValueOnce({
        categories: Array.from({ length: 50 }, (_, index) => ({ id: `c-${100 + index}` })),
        total: 150,
        hasMore: false,
      });

    const categories = await fetchAllCategories();

    expect(categories).toHaveLength(150);
    expect(categoryClient.listCategories).toHaveBeenNthCalledWith(1, {
      limit: 100,
      offset: 0,
    });
    expect(categoryClient.listCategories).toHaveBeenNthCalledWith(2, {
      limit: 100,
      offset: 100,
    });
  });
});
