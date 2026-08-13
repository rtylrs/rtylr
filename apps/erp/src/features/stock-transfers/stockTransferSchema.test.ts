import { describe, expect, it } from 'vitest';
import { toStockTransferItems } from './stockTransferSchema';

describe('toStockTransferItems', () => {
  it('builds positive transfer lines from the modal draft', () => {
    expect(
      toStockTransferItems([
        { productId: '12', quantity: '2.5' },
        { productId: '34', quantity: '1' },
      ])
    ).toEqual([
      { productId: '12', quantity: 2.5 },
      { productId: '34', quantity: 1 },
    ]);
  });

  it('drops incomplete and non-positive lines so empty transfers cannot be submitted', () => {
    expect(
      toStockTransferItems([
        { productId: '', quantity: '2' },
        { productId: '12', quantity: '' },
        { productId: '34', quantity: '0' },
        { productId: '56', quantity: '-1' },
        { productId: '78', quantity: '3' },
      ])
    ).toEqual([{ productId: '78', quantity: 3 }]);
  });
});
