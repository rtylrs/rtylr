import { describe, expect, it } from 'vitest';
import { stockTakeProgress } from './stockTakeSchema';

describe('stockTakeProgress', () => {
  it('counts a valid zero quantity as counted', () => {
    expect(stockTakeProgress([{ countedQty: 0 }])).toEqual({
      counted: 1,
      total: 1,
      pct: 100,
    });
  });

  it('reports an empty list as zero progress', () => {
    expect(stockTakeProgress([])).toEqual({ counted: 0, total: 0, pct: 0 });
  });
});
