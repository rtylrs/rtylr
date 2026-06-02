import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/v1/purchase_orders';
import { erpApi, PurchaseOrder } from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

interface UsePurchaseOrdersParams {
  page?: number;
  pageSize?: number;
}

export interface PurchaseOrdersResult {
  orders: PurchaseOrder[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

// Backend (erp.v1.PurchaseOrder) shape, camelCased by @voxire/api on the way in.
// The backend never sends a supplier name, a `date`, an `expectedDate`, or an
// item count — it sends supplierId, expectedDelivery, and an items[] array.
// Status is a lowercase enum: draft|sent|approved|partial|received|cancelled.
const purchaseOrderItemSchema = z.object({
  id: z.number(),
  purchaseOrderId: z.number().default(0),
  productId: z.number().default(0),
  quantityOrdered: z.number().default(0),
  quantityReceived: z.number().default(0),
  unitCost: z.number().default(0),
  totalCost: z.number().default(0),
});

const backendPurchaseOrderSchema = z.object({
  id: z.number(),
  supplierId: z.number().default(0),
  poNumber: z.string().default(''),
  status: z.string().default('draft'),
  subtotal: z.number().default(0),
  taxAmount: z.number().default(0),
  total: z.number().default(0),
  expectedDelivery: z.string().optional().nullable(),
  receivedAt: z.string().optional().nullable(),
  notes: z.string().default(''),
  createdAt: z.string().default(''),
  updatedAt: z.string().default(''),
  items: z.array(purchaseOrderItemSchema).default([]),
});

type BackendPurchaseOrder = z.infer<typeof backendPurchaseOrderSchema>;

// Map the backend's lowercase status enum onto the display-cased union the
// shared PurchaseOrder type / UI components expect. `partial` is surfaced as
// `Sent` so partially-received orders still appear in the open/receiving flow;
// `cancelled` has no display equivalent and falls back to `Draft`.
function toDisplayStatus(status: string): PurchaseOrder['status'] {
  switch (status) {
    case 'sent':
      return 'Sent';
    case 'approved':
      return 'Approved';
    case 'partial':
      return 'Sent';
    case 'received':
      return 'Received';
    case 'draft':
    case 'cancelled':
    default:
      return 'Draft';
  }
}

function adaptPurchaseOrder(po: BackendPurchaseOrder): PurchaseOrder {
  return {
    id: po.id,
    // Backend does not return a supplier name; degrade to a stable label.
    supplier: po.supplierId > 0 ? `Supplier #${po.supplierId}` : '—',
    supplierId: po.supplierId,
    // No standalone `date` field exists; use the created timestamp.
    date: po.createdAt,
    expectedDate: po.expectedDelivery ?? '',
    // `items` is an array on the backend; the UI expects a line-item count.
    items: po.items.length,
    total: po.total.toString(),
    status: toDisplayStatus(po.status),
  };
}

export function usePurchaseOrders(params: UsePurchaseOrdersParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<PurchaseOrdersResult>({
    queryKey: ['purchase-orders', { page, pageSize }],
    queryFn: async (): Promise<PurchaseOrdersResult> => {
      const res = await unwrap(erpApi.listPurchaseOrders({ page, page_size: pageSize }));
      const raw = res as unknown as Record<string, unknown>;
      // Backend envelope key `purchase_orders` arrives camelCased as `purchaseOrders`.
      const orders = z
        .array(backendPurchaseOrderSchema)
        .parse(unwrapList(res, 'purchaseOrders'))
        .map(adaptPurchaseOrder);
      return {
        orders,
        hasMore: Boolean(raw.hasMore),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.pageSize ?? pageSize),
      };
    },
  });
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePurchaseOrderRequest) => unwrap(erpApi.createPurchaseOrder(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to create purchase order'),
  });
}

export function useUpdatePurchaseOrderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      unwrap(erpApi.updatePurchaseOrderStatus(id, status)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to update purchase order status'),
  });
}

export function useApprovePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.approvePurchaseOrder(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to approve purchase order'),
  });
}

export function useSendPurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    // Send transitions a draft PO into 'sent' status. Backend hangs
    // off updatePurchaseOrderStatus for now - when a dedicated
    // /send endpoint is added (with the supplier email side-effect),
    // swap the mutationFn here and the call sites stay unchanged.
    // Status value must be the backend's lowercase enum.
    mutationFn: (id: number) => unwrap(erpApi.updatePurchaseOrderStatus(id, 'sent')),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to send purchase order'),
  });
}
