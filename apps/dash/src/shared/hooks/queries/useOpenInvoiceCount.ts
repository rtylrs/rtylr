import { useQuery } from '@tanstack/react-query';
import { toFiniteNumber } from '@voxire/api';
import { insightsFinancialClient } from '../../lib/clients';

/**
 * How many invoices are outstanding.
 *
 * hub.dashboard.v1.GetKPIsResponse has no invoice count, which is why the
 * Invoices card used to read a field that did not exist and render a confident 0
 * forever (rtylr#1001).
 *
 * insights.financial.v1 getInvoiceAging already provides it, and dash already
 * calls that endpoint for InvoiceAgingWidget. Aging buckets are outstanding
 * invoices by definition - current, 1-30, 31-60, 61-90, 90+ - so their counts
 * sum to the number of open invoices. Nothing new is needed server-side.
 *
 * toFiniteNumber because OneKit carries int64 as a string on the wire, and both
 * Number('') and Number(null) are 0-or-NaN traps.
 */
export const useOpenInvoiceCount = () =>
  useQuery({
    queryKey: ['insights-open-invoice-count'],
    queryFn: async (): Promise<number> => {
      const res = await insightsFinancialClient.getInvoiceAging({});
      return (res.aging?.buckets ?? []).reduce(
        (total, bucket) => total + toFiniteNumber(bucket.count),
        0
      );
    },
    staleTime: 60_000,
  });
