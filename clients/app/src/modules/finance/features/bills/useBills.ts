import { createOneKitClient } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { BillServiceClient } from '@voxire/voxie/finance/bill/v1/client';
import type { Bill as ApiBill } from '@voxire/voxie/finance/bill/v1/types';
import { billSchema } from '../../shared/schemas/bills';

const billClient = createOneKitClient(BillServiceClient, 'finance');

export const BILLS_PAGE_SIZE = 50;

function toBillView(b: ApiBill): unknown {
  // vendor is supplier_name, which the server resolves and the contract marks
  // @required. Keep description separate: UpdateBillBody requires it when a
  // bill is marked paid, so it must not be replaced with the vendor name.
  return {
    id: b.id,
    supplier: b.supplierId,
    vendor: b.supplierName,
    supplierName: b.supplierName,
    currency: b.currency,
    description: b.description,
    amount: b.amount,
    status: b.status,
    dueDate: b.dueDate,
  };
}

/**
 * Fetches the complete bill set in pages.
 *
 * The bills endpoint is paginated (50 rows by default), so treating its first
 * response as the complete list silently hides later bills and makes the
 * outstanding totals depend on page one. Loading every page here lets the
 * screen page locally while keeping its metrics and bulk actions global.
 */
export function useBills() {
  return useQuery({
    queryKey: ['finance', 'bills'],
    queryFn: async () => {
      const allBills: unknown[] = [];
      let page = 1;

      while (true) {
        const res = await billClient.listBills({ page, pageSize: BILLS_PAGE_SIZE });
        const pageBills = res.bills ?? [];
        allBills.push(...pageBills.map(toBillView));

        if (pageBills.length === 0 || res.hasMore === false) break;

        // Older compatibility responses may omit has_more. A short page is
        // conclusive in that case; a full page warrants one more request.
        if (res.hasMore === undefined && pageBills.length < BILLS_PAGE_SIZE) break;

        page += 1;
      }

      return z.array(billSchema).parse(allBills);
    },
  });
}
