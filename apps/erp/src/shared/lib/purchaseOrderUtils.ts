/**
 * Purchase order helper functions used across the purchasing feature.
 */
import type { PurchaseOrder } from '../services/erpApi';

/** Return true when a purchase order is in a terminal state. */
export function isPurchaseOrderTerminal(order: PurchaseOrder): boolean {
  return order.status === 'Received';
}

/** Return true when a purchase order is editable (still a draft). */
export function isPurchaseOrderEditable(order: PurchaseOrder): boolean {
  return order.status === 'Draft';
}

/** Return true when a purchase order can be approved. */
export function canApprovePurchaseOrder(order: PurchaseOrder): boolean {
  return order.status === 'Pending Approval';
}

/** Return true when a purchase order can be sent to the supplier. */
export function canSendPurchaseOrder(order: PurchaseOrder): boolean {
  return order.status === 'Draft' || order.status === 'Approved';
}

/** Return a badge variant for a purchase order status. */
export function purchaseOrderStatusVariant(
  order: PurchaseOrder
): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (order.status) {
    case 'Approved':
    case 'Received':
      return 'default';
    case 'Sent':
      return 'secondary';
    case 'Draft':
      return 'outline';
    case 'Pending Approval':
      return 'destructive';
    default:
      return 'outline';
  }
}

/** Sort purchase orders by date, newest first. */
export function sortOrdersByDateDesc(orders: PurchaseOrder[]): PurchaseOrder[] {
  return [...orders].sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
  );
}

/** Filter orders by a given status. */
export function filterOrdersByStatus(
  orders: PurchaseOrder[],
  status: PurchaseOrder['status']
): PurchaseOrder[] {
  return orders.filter((o) => o.status === status);
}

/** Return the number of draft orders in a list. */
export function countDraftOrders(orders: PurchaseOrder[]): number {
  return orders.filter((o) => o.status === 'Draft').length;
}

/** Return the number of orders pending approval in a list. */
export function countPendingApprovalOrders(orders: PurchaseOrder[]): number {
  return orders.filter((o) => o.status === 'Pending Approval').length;
}
