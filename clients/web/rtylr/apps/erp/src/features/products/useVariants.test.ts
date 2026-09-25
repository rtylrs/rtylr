import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fetchAllVariants } from './useVariants';

const mocks = vi.hoisted(() => ({
  listVariants: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ listVariants: mocks.listVariants }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

describe('fetchAllVariants', () => {
  beforeEach(() => {
    mocks.listVariants.mockReset();
  });

  it('follows server pagination instead of treating the first 100 rows as complete', async () => {
    mocks.listVariants
      .mockResolvedValueOnce({
        variants: Array.from({ length: 100 }, (_, index) => ({ id: String(index + 1) })),
        hasMore: true,
        total: 101,
      })
      .mockResolvedValueOnce({
        variants: [{ id: '101' }],
        hasMore: false,
        total: 101,
      });

    await expect(fetchAllVariants()).resolves.toHaveLength(101);
    expect(mocks.listVariants).toHaveBeenNthCalledWith(1, {
      productId: '0',
      limit: 100,
      offset: 0,
    });
    expect(mocks.listVariants).toHaveBeenNthCalledWith(2, {
      productId: '0',
      limit: 100,
      offset: 100,
    });
  });

  it('fails instead of looping when the server repeats a page', async () => {
    mocks.listVariants.mockResolvedValue({
      variants: [{ id: '1' }],
      hasMore: true,
      total: 2,
    });

    await expect(fetchAllVariants('5201')).rejects.toThrow('Variant pagination did not advance');
  });
});
