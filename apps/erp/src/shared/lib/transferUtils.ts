/**
 * Stock transfer helper functions used across the stock-transfers feature.
 */
import type { StockTransfer } from '../services/erpApi';

/** Return true when a stock transfer is in a terminal state. */
export function isTransferComplete(transfer: StockTransfer): boolean {
  return transfer.status === 'Received' || transfer.status === 'Cancelled';
}

/** Return true when a transfer is currently in transit. */
export function isTransferInTransit(transfer: StockTransfer): boolean {
  return transfer.status === 'In Transit';
}

/** Return true when a transfer has been cancelled. */
export function isTransferCancelled(transfer: StockTransfer): boolean {
  return transfer.status === 'Cancelled';
}

/** Return a badge variant for a stock transfer status. */
export function transferStatusVariant(
  transfer: StockTransfer
): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (transfer.status) {
    case 'Received':
      return 'default';
    case 'In Transit':
      return 'secondary';
    case 'Cancelled':
      return 'destructive';
    case 'Draft':
      return 'outline';
    default:
      return 'outline';
  }
}

/** Sort transfers by creation date, newest first. */
export function sortTransfersByDateDesc(transfers: StockTransfer[]): StockTransfer[] {
  return [...transfers].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Filter transfers by status. */
export function filterTransfersByStatus(
  transfers: StockTransfer[],
  status: StockTransfer['status']
): StockTransfer[] {
  return transfers.filter((t) => t.status === status);
}

/** Return the total item count across all transfers. */
export function totalItemsInTransfer(transfers: StockTransfer[]): number {
  return transfers.reduce((sum, t) => sum + t.items, 0);
}
