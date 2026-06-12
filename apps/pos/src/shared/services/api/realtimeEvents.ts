/**
 * Typed definitions for every realtime event emitted over the POS WebSocket.
 * The backend (onion) emits these on topic `owner:{userId}`.
 */

// ── Order events ──────────────────────────────────────────────────────────────

export interface PosOrderCreatedEvent {
  type: 'pos:order:created';
  payload: {
    orderId: number;
    orderNumber: string;
    locationId: number;
    tableId: number | null;
    status: 'open';
    total: number;
    createdAt: string;
  };
}

export interface PosOrderUpdatedEvent {
  type: 'pos:order:updated';
  payload: {
    orderId: number;
    status: string;
    updatedAt: string;
  };
}

export interface PosOrderPaidEvent {
  type: 'pos:order:paid';
  payload: {
    orderId: number;
    orderNumber: string;
    total: number;
    paymentMethod: string;
    paidAt: string;
  };
}

export interface PosOrderVoidedEvent {
  type: 'pos:order:voided';
  payload: {
    orderId: number;
    reason: string;
    voidedAt: string;
  };
}

// ── KDS events ────────────────────────────────────────────────────────────────

export interface PosOrderPreparingEvent {
  type: 'pos:order:preparing';
  payload: {
    orderId: number;
    station: string;
    startedAt: string;
  };
}

export interface PosKdsItemUpdatedEvent {
  type: 'pos:kds:item:updated';
  payload: {
    itemId: number;
    orderId: number;
    status: 'new' | 'cooking' | 'ready' | 'served';
    station: string;
    updatedAt: string;
  };
}

// ── Table events ──────────────────────────────────────────────────────────────

export interface PosTableStatusEvent {
  type: 'pos:table:status';
  payload: {
    tableId: number;
    status: 'available' | 'occupied' | 'reserved';
    orderId: number | null;
    updatedAt: string;
  };
}

// ── Session events ────────────────────────────────────────────────────────────

export interface PosSessionOpenedEvent {
  type: 'pos:session:opened';
  payload: {
    sessionId: number;
    terminalId: number;
    openedAt: string;
  };
}

export interface PosSessionClosedEvent {
  type: 'pos:session:closed';
  payload: {
    sessionId: number;
    closedAt: string;
    netSales: number;
    orderCount: number;
  };
}

// ── Inventory events ──────────────────────────────────────────────────────────

export interface PosStockLowEvent {
  type: 'pos:stock:low';
  payload: {
    productId: number;
    productName: string;
    currentQty: number;
    threshold: number;
  };
}

export interface PosStockOutEvent {
  type: 'pos:stock:out';
  payload: {
    productId: number;
    productName: string;
  };
}

// ── Union type ────────────────────────────────────────────────────────────────

export type PosRealtimeEvent =
  | PosOrderCreatedEvent
  | PosOrderUpdatedEvent
  | PosOrderPaidEvent
  | PosOrderVoidedEvent
  | PosOrderPreparingEvent
  | PosKdsItemUpdatedEvent
  | PosTableStatusEvent
  | PosSessionOpenedEvent
  | PosSessionClosedEvent
  | PosStockLowEvent
  | PosStockOutEvent;

export type PosEventType = PosRealtimeEvent['type'];

/** Type guard: narrow a generic realtime event to a specific POS event. */
export function isPosEvent<T extends PosRealtimeEvent>(
  event: { type: string },
  type: T['type'],
): event is T {
  return event.type === type;
}

/** All KDS-relevant event types - used by useKdsStream to filter the bus. */
export const KDS_EVENT_TYPES = new Set<PosEventType>([
  'pos:order:created',
  'pos:order:preparing',
  'pos:kds:item:updated',
]);
