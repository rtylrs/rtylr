import { describe, expect, it, vi } from 'vitest';
import { importProducts, toCreateProductInput } from './importProducts';

const row = {
  name: 'Espresso Beans',
  sku: 'ESP-01',
  category_id: '2',
  supplier_id: '7',
  stock_quantity: '12',
  sale_price: '25.50',
};

describe('product CSV import', () => {
  it('maps CSV values to the generated create-product request', () => {
    expect(toCreateProductInput(row, 2)).toMatchObject({
      categoryId: '2',
      supplierId: '7',
      name: 'Espresso Beans',
      sku: 'ESP-01',
      barcode: 'ESP-01',
      description: 'Espresso Beans',
      unit: 'pcs',
      stockQuantity: 12,
      salePrice: 25.5,
    });
  });

  it('persists every row and reports the number created', async () => {
    const createProduct = vi.fn().mockResolvedValue({ productId: '101' });

    await expect(importProducts([row, { ...row, sku: 'ESP-02' }], createProduct)).resolves.toBe(2);
    expect(createProduct).toHaveBeenCalledTimes(2);
    expect(createProduct).toHaveBeenNthCalledWith(2, expect.objectContaining({ sku: 'ESP-02' }));
  });

  it('rejects a row before calling the API when a required ID is invalid', async () => {
    const createProduct = vi.fn();

    await expect(
      importProducts([{ ...row, supplier_id: 'North Roast' }], createProduct)
    ).rejects.toThrow('Row 2: supplier_id must be a positive numeric ID');
    expect(createProduct).not.toHaveBeenCalled();
  });
});
