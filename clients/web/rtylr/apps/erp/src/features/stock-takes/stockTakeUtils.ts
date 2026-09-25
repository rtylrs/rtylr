import type { InventoryReportItem } from '../inventory/useInventory';
import type { CreateStockTakeItemRequest } from '@voxire/voxie/erp/stock_take/v1/types';

export function buildStockTakeItems(
  inventoryItems: Array<Pick<InventoryReportItem, 'productId' | 'stockQuantity'>>
): CreateStockTakeItemRequest[] {
  return inventoryItems.flatMap(({ productId, stockQuantity }) => {
    if (!productId || Number(productId) <= 0) return [];

    return [
      {
        productId,
        expectedQty: Math.max(0, stockQuantity ?? 0),
        countedQty: 0,
      },
    ];
  });
}
