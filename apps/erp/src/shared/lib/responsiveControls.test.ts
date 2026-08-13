import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');

function source(path: string): string {
  return readFileSync(join(sourceRoot, path), 'utf8');
}

describe('ERP responsive controls', () => {
  it('keeps listed touch controls at the 44px minimum', () => {
    const receiving = source('features/purchasing/Receiving.tsx');
    const quickAdjust = source('shared/components/QuickStockAdjust.tsx');
    const layout = source('layouts/DefaultLayout.tsx');

    expect(receiving).toContain('className="h-11 w-11 rounded flex items-center justify-center"');
    expect(receiving).toContain(
      'className="h-11 w-11 rounded-lg flex items-center justify-center"'
    );
    expect(quickAdjust).toContain('className="h-11 w-11 rounded flex items-center justify-center"');
    expect(layout).toContain('className="h-11 w-11 rounded-lg flex items-center justify-center"');
    expect(layout).not.toContain('className="h-8 w-8 rounded-lg flex items-center justify-center"');
    expect(layout).not.toContain('className="h-7 w-7 rounded-lg flex items-center justify-center"');
  });

  it('uses mobile-safe input sizing in stock transfer and stock take forms', () => {
    expect(source('features/stock-transfers/StockTransfers.tsx')).toContain("fontSize: '16px'");
    expect(source('features/stock-takes/StockTakes.tsx')).toContain("fontSize: '16px'");
  });
});
