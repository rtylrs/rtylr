import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { PurchaseOrderServiceClient } from '@voxire/voxie/erp/purchase_order/v1/client';

// ERP is a foreign domain: read-only, caller-scoped client id.
const poClient = createOneKitClient(PurchaseOrderServiceClient, 'finance');

const receivedPoSchema = z.object({
  id: z.string(),
  supplierId: z.string().optional(),
  poNumber: z.string().optional(),
  total: z.number().optional(),
  quantityReceived: z.number().optional(),
});

export interface ReceivedPurchaseOrder {
  id: string;
  supplierId?: string;
  poNumber?: string;
  total: number;
}

export const ERP_RECEIVED_POS_QUERY_KEY = ['erp', 'purchase-orders', 'received'] as const;

/**
 * Purchase orders received into the warehouse: the payables commitment that
 * has happened but is not yet in this business's accounts payable (ERP->FIN-03).
 *
 * Honesty note (#850): per-PO "settled by which bill" needs a bill-to-PO
 * link, which the finance.bill.v1 contract does not carry yet (drafted on
 * voxie branch docs/cross-app-round2). Until that lands, this reports the
 * received set itself, which is the commitment figure; it does not claim
 * per-order billing status.
 *
 * Degrade: a tenant without ERP yields null and the section hides. Never
 * render an empty-zero payables commitment.
 */
export function useErpReceivedPurchaseOrders(options?: { enabled?: boolean }) {
  return useQuery<ReceivedPurchaseOrder[] | null>({
    queryKey: ERP_RECEIVED_POS_QUERY_KEY,
    queryFn: async (): Promise<ReceivedPurchaseOrder[] | null> => {
      try {
        const res = await poClient.listPurchaseOrders({
          status: 'received',
          page: 1,
          pageSize: 100,
        });
        const pos = (res.purchaseOrders ?? [])
          .map((po) => receivedPoSchema.parse(po))
          .filter((po) => (po.quantityReceived ?? 0) > 0)
          .map((po) => ({
            id: po.id,
            supplierId: po.supplierId,
            poNumber: po.poNumber,
            total: po.total ?? 0,
          }));
        return pos;
      } catch {
        // No ERP for this tenant, or the read failed: hide rather than zero.
        return null;
      }
    },
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
