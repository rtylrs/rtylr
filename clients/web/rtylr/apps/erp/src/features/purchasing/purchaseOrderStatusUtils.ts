export type PurchaseOrderStatusCard = 'Draft' | 'Sent' | 'Approved' | 'Received';

export type PurchaseOrderFilter = PurchaseOrderStatusCard | 'All';

export type PurchaseOrderDisplayStatus = PurchaseOrderStatusCard | 'Partial' | 'Cancelled';

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
    Received: purchaseOrders.filter(
      (order) => displayPurchaseOrderStatus(order.status ?? '') === 'Received'
    ).length,
  };
}
