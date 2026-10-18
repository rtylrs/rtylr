export type PurchaseOrderStatusCard = 'Draft' | 'Sent' | 'Approved' | 'Partial' | 'Received';

export type PurchaseOrderFilter = PurchaseOrderStatusCard | 'All';

export type PurchaseOrderDisplayStatus = PurchaseOrderStatusCard | 'Cancelled';

export function displayPurchaseOrderStatus(status: string): PurchaseOrderDisplayStatus {
  switch (status) {
    case 'sent':
      return 'Sent';
    case 'approved':
      return 'Approved';
    case 'received':
      return 'Received';
    case 'partial':
      return 'Partial';
    case 'cancelled':
      return 'Cancelled';
    case 'draft':
    default:
      return 'Draft';
  }
}

export function purchaseOrderStatusFilter(filter: PurchaseOrderFilter): string {
  return filter === 'All' ? '' : filter.toLowerCase();
}

export function getPurchaseOrderStatusCounts(
  purchaseOrders: ReadonlyArray<{ status?: string | null }>
): Record<PurchaseOrderStatusCard, number> {
  return {
    Draft: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Draft'
    ).length,
    Sent: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Sent'
    ).length,
    Approved: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Approved'
    ).length,
    Partial: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Partial'
    ).length,
    Received: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Received'
    ).length,
  };
}

/** The number the business gave the order, or its id when it has none. */
export function purchaseOrderLabel(order: { id?: string; poNumber?: string }): string {
  return order.poNumber?.trim() || `PO-${order.id ?? ''}`;
}
