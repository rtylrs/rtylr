import { describe, expect, it } from 'vitest';
import { buildStockTakeItems } from './stockTakeUtils';

describe('buildStockTakeItems', () => {
  it('creates an uncounted line with current stock as the expected quantity', () => {
    expect(
      buildStockTakeItems([
        { productId: '12', stockQuantity: 7 },
        { productId: '34', stockQuantity: 0 },
      ])
    ).toEqual([
      { productId: '12', expectedQty: 7, countedQty: 0 },
      { productId: '34', expectedQty: 0, countedQty: 0 },
    ]);
  });

  it('ignores inventory rows without a valid product and clamps invalid stock', () => {
    expect(
      buildStockTakeItems([
        { productId: undefined, stockQuantity: 4 },
        { productId: '0', stockQuantity: 4 },
        { productId: '56', stockQuantity: -3 },
        { productId: '78', stockQuantity: undefined },
      ])
    ).toEqual([
      { productId: '56', expectedQty: 0, countedQty: 0 },
      { productId: '78', expectedQty: 0, countedQty: 0 },
    ]);
  });
});
