import { describe, expect, it, vi } from 'vitest';
import { BrandServiceClient } from '@voxire/voxie/erp/brand/v1/client';
import { ERP_LIST_LIMIT, ERP_LIST_OFFSET, withErpListBounds } from './erpListBounds';

describe('withErpListBounds', () => {
  it('adds the shared bounded window without dropping endpoint filters', () => {
    expect(withErpListBounds({ productId: '42', daysAhead: 30 })).toEqual({
      productId: '42',
      daysAhead: 30,
      limit: ERP_LIST_LIMIT,
      offset: ERP_LIST_OFFSET,
    });
  });

  it('keeps every ERP catalog/report request within the maximum page size', () => {
    expect(ERP_LIST_LIMIT).toBe(100);
    expect(ERP_LIST_OFFSET).toBe(0);
  });

  it('serializes the bound into the generated catalog endpoint query', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ brands: [] }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      })
    );
    const client = new BrandServiceClient('https://api.example', {
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.listBrands(withErpListBounds({}));

    const requestUrl = String(fetchMock.mock.calls[0]?.[0]);
    expect(new URL(requestUrl).pathname).toBe('/erp/brand/v1/brands');
  });
});
