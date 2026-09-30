import { createOneKitClient } from '@voxire/api';
import { StockServiceClient } from '@voxire/voxie/pos/stock/v1/client';
import { submitStockAction, type BuiltStockAction, type StockActionSubmission } from './inventoryWorkflow';

const stockClient = createOneKitClient(StockServiceClient, 'pos');

export const listBranchInventory = (branchId: string) =>
  stockClient.listInventory({ branchId });

export const listLowStock = (branchId: string) => stockClient.lowStock({ branchId });

/**
 * Sends one stock action to the service that owns it.
 *
 * The transfer route requires an `Idempotency-Key` (the server answers 428
 * without it), so it is attached here, from the key `submitStockAction` minted
 * once for the operator action. It travels in `RequestOptions`, which the
 * generated client already forwards, so no contract change is needed.
 *
 * Takes a submission rather than a built action, so a retry of one action
 * re-sends the key that action already carries instead of minting a second one.
 */
export function sendStockAction({ action, payload, idempotencyKey }: StockActionSubmission) {
  switch (action) {
    case 'receive':
      return stockClient.receiveStock(payload);
    case 'adjust':
      return stockClient.adjustStock(payload);
    case 'waste':
      return stockClient.wasteStock(payload);
    case 'transfer':
      return stockClient.transferStock(payload, {
        headers: { 'Idempotency-Key': idempotencyKey },
      });
  }
}

/** Mints this operator action's key and sends it. */
export function recordStockAction(built: BuiltStockAction): Promise<unknown> {
  return submitStockAction(built, sendStockAction);
}
