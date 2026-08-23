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
      // vendor is supplier_name, which the server resolves and the contract
      // marks @required. The comment that used to sit here claimed the Bill
      // message had no vendor name and that description carried it instead;
      // that was true once and the mapping outlived it, so the column showed a
      // free-text description where a vendor belonged.
      //
      // description stays separate rather than doubling as the vendor: it is
      // @required on UpdateBillBody, so mark-as-paid sends it back, and
      // conflating the two would write the vendor name into it.
      return z.array(billSchema).parse(
        (res.bills ?? []).map((b) => ({
          id: b.id,
          supplier: b.supplierId,
          vendor: b.supplierName,
          supplierName: b.supplierName,
          currency: b.currency,
          description: b.description,
          amount: b.amount,
          status: b.status,
          dueDate: b.dueDate,
        }))
      );
    },
  });
}
