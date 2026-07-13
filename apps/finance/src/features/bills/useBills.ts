import { createOneKitClient } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { BillServiceClient } from '@voxire/voxie/finance/bill/v1/client';
import { billSchema } from '../../shared/schemas/bills';

const billClient = createOneKitClient(BillServiceClient, 'finance');

export function useBills() {
  return useQuery({
    queryKey: ['bills'],
    queryFn: async () => {
      const res = await billClient.listBills({});
      // finance.bill.v1.Bill has no vendor name or category - only a
      // supplier_id FK and a free-text description (which the create form
      // already uses to carry the vendor name when the field is left blank).
      return z.array(billSchema).parse(
        (res.bills ?? []).map((b) => ({
          id: b.id,
          supplier: b.supplierId,
          vendor: b.description,
          amount: b.amount,
          status: b.status,
          dueDate: b.dueDate,
        }))
      );
    },
  });
}
