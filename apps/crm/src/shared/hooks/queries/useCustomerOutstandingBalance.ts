import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';

// Finance is a foreign domain: read-only lookup so the CRM customer page can
// show outstanding balance. Client app id stays 'crm'.
const invoicesClient = createOneKitClient(InvoicesServiceClient, 'crm');

export interface CustomerBalance {
  totalOutstanding: number;
  invoiceCount: number;
  currency?: string;
}

export const CUSTOMER_BALANCE_KEY = (customerId: string) =>
  ['finance', 'invoices', 'by-customer', customerId] as const;

/**
 * Outstanding balance for one CRM customer across all non-draft, non-void
 * invoices carrying their customer_id (#841). Served off the
 * (business_id, customer_id) index added in onion#879.
 *
 * Degrade: a tenant without Finance — or a failed read — yields null and the
 * panel hides; never an empty-zero claim about someone's balance.
 */
export function useCustomerOutstandingBalance(customerId?: string) {
  return useQuery<CustomerBalance | null>({
    queryKey: CUSTOMER_BALANCE_KEY(customerId ?? ''),
    queryFn: async (): Promise<CustomerBalance | null> => {
      if (!customerId) return null;
      try {
        const res = await invoicesClient.listInvoices({
          customerId,
          pageSize: 500,
          page: 1,
        });
        if (res === null || res === undefined) return null;
        const invoices = res.invoices ?? [];
        const billable = invoices.filter(
          (inv) => inv.status !== 'draft' && inv.status !== 'void' && inv.status !== 'cancelled'
        );
        if (!billable.length) return { totalOutstanding: 0, invoiceCount: 0 };
        const totalOutstanding = billable.reduce((sum, inv) => sum + (inv.amountDue ?? 0), 0);
        return {
          totalOutstanding,
          invoiceCount: billable.length,
        };
      } catch {
        return null;
      }
    },
    enabled: !!customerId,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
