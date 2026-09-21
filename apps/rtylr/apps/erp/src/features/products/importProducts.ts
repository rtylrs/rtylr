import type { CreateProductRequest } from '@voxire/voxie/erp/product/v1/types';

export type ProductImportRow = Record<string, string>;
export type CreateProduct = (input: CreateProductRequest) => Promise<unknown>;

function value(row: ProductImportRow, ...keys: string[]): string {
  for (const key of keys) {
    const candidate = row[key]?.trim();
    if (candidate) return candidate;
  }
  return '';
}

function required(
  row: ProductImportRow,
  rowNumber: number,
  label: string,
  ...keys: string[]
): string {
  const result = value(row, ...keys);
  if (!result) throw new Error(`Row ${rowNumber}: ${label} is required`);
  return result;
}

function id(row: ProductImportRow, rowNumber: number, label: string, ...keys: string[]): string {
  const result = required(row, rowNumber, label, ...keys);
  if (!/^\d+$/.test(result) || Number(result) <= 0) {
    throw new Error(`Row ${rowNumber}: ${label} must be a positive numeric ID`);
  }
  return result;
}

function nonNegativeNumber(
  row: ProductImportRow,
  rowNumber: number,
  label: string,
  ...keys: string[]
): number {
  const result = value(row, ...keys);
  if (!result) return 0;
  const parsed = Number(result);
  if (!Number.isFinite(parsed) || parsed < 0) {
    throw new Error(`Row ${rowNumber}: ${label} must be a non-negative number`);
  }
  return parsed;
}

function booleanValue(row: ProductImportRow, ...keys: string[]): boolean | undefined {
  const result = value(row, ...keys).toLowerCase();
  if (!result) return undefined;
  if (['true', '1', 'yes'].includes(result)) return true;
  if (['false', '0', 'no'].includes(result)) return false;
  return undefined;
}

export function toCreateProductInput(
  row: ProductImportRow,
  rowNumber: number
): CreateProductRequest {
  const name = required(row, rowNumber, 'name', 'name');
  const sku = required(row, rowNumber, 'sku', 'sku');

  return {
    categoryId: id(row, rowNumber, 'category_id', 'category_id', 'category'),
    supplierId: id(row, rowNumber, 'supplier_id', 'supplier_id', 'supplier'),
    name,
    sku,
    barcode: value(row, 'barcode') || sku,
    description: value(row, 'description') || name,
    unit: value(row, 'unit') || 'pcs',
    stockQuantity: nonNegativeNumber(row, rowNumber, 'stock_quantity', 'stock_quantity', 'stock'),
    costPrice: nonNegativeNumber(row, rowNumber, 'cost_price', 'cost_price', 'costPrice'),
    salePrice: nonNegativeNumber(row, rowNumber, 'sale_price', 'sale_price', 'salePrice'),
    lowStockThreshold: nonNegativeNumber(
      row,
      rowNumber,
      'low_stock_threshold',
      'low_stock_threshold',
      'lowStock'
    ),
    reorderQuantity: nonNegativeNumber(
      row,
      rowNumber,
      'reorder_quantity',
      'reorder_quantity',
      'reorderPoint'
    ),
    isActive: booleanValue(row, 'is_active') ?? true,
    trackInventory: booleanValue(row, 'track_inventory') ?? true,
    currency: value(row, 'currency') || 'USD',
  };
}

export async function importProducts(
  rows: ProductImportRow[],
  createProduct: CreateProduct
): Promise<number> {
  // Validate every row before creating anything so one bad row cannot leave earlier rows behind.
  const inputs = rows.map((row, index) => toCreateProductInput(row, index + 2));

  let created = 0;
  for (const input of inputs) {
    await createProduct(input);
    created += 1;
  }
  return created;
}
