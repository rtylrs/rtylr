import { describe, expect, it, vi } from 'vitest';
import { invalidateStockDerivedQueries, STOCK_DERIVED_QUERY_KEYS } from './stockInvalidation';

describe('invalidateStockDerivedQueries', () => {
  it('invalidates every view that derives data from stock', () => {
    const invalidateQueries = vi.fn().mockResolvedValue(undefined);

    invalidateStockDerivedQueries({ invalidateQueries });

    expect(invalidateQueries).toHaveBeenCalledTimes(STOCK_DERIVED_QUERY_KEYS.length);
    expect(invalidateQueries.mock.calls.map(([filters]) => filters?.queryKey)).toEqual(
      STOCK_DERIVED_QUERY_KEYS
    );
  });
});
