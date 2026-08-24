import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { BillServiceClient } from '@voxire/voxie/finance/bill/v1/client';

// Finance is a foreign domain: read-only lookup so the ERP supplier page can
// show outstanding bills. Client app id stays 'erp'.
const billClient = createOneKitClient(BillServiceClient, 'erp');

export interface SupplierBill {
  id?: string;
  status?: string;
  amount?: number;
  dueDate?: string;
  currency?: string;
}

export const SUPPLIER_BILLS_KEY = (supplierId: string) =>
  ['finance', 'bills', 'by-supplier', supplierId] as const;

/**
 * Bills for one supplier (#844). Served off finance.bill.v1.listBills with
 * the supplier_id filter.
 *
 * Degrade: a tenant without Finance — or a failed read — yields null and the
 * panel hides; never an empty-zero claim about someone's payment history.
 */
export function useSupplierBills(supplierId: string | undefined) {
  return useQuery<SupplierBill[] | null>({
    queryKey: SUPPLIER_BILLS_KEY(supplierId ?? ''),
    queryFn: async (): Promise<SupplierBill[] | null> => {
      if (!supplierId) return null;
      try {
        const res = await billClient.listBills({
          supplierId,
          page: 1,
          pageSize: 50,
        });
        if (res === null || res === undefined) return null;
        return (res.bills ?? []).map((b) => ({
          id: b.id,
          status: b.status,
          amount: b.amount,
          dueDate: b.dueDate,
          currency: b.currency,
        }));
      } catch {
        return null;
      }
    },
    enabled: !!supplierId,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
