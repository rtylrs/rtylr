import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';

// Finance is a foreign domain: read-only lookup so the CRM customer page can
// show outstanding balance. Client app id stays 'crm'.
const invoicesClient = createOneKitClient(InvoicesServiceClient, 'crm');
const INVOICE_PAGE_SIZE = 500;
const MAX_INVOICE_PAGES = 100;

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
        const seenInvoices = new Set<string>();
        let totalOutstanding = 0;
        let invoiceCount = 0;
        let currency: string | undefined;

        for (let page = 1; page <= MAX_INVOICE_PAGES; page += 1) {
          const res = await invoicesClient.listInvoices({
            customerId,
            pageSize: INVOICE_PAGE_SIZE,
            page,
          });
          if (res === null || res === undefined) return null;

          const invoices = res.invoices ?? [];
          let newInvoiceCount = 0;
          for (const invoice of invoices) {
            const key = invoice.id ?? invoice.invoiceNumber;
            if (key !== undefined) {
              if (seenInvoices.has(key)) continue;
              seenInvoices.add(key);
            }
            newInvoiceCount += 1;
            if (
              invoice.status === 'draft' ||
              invoice.status === 'void' ||
              invoice.status === 'cancelled'
            ) {
              continue;
            }
            if (currency && invoice.currency && currency !== invoice.currency) return null;
            currency ||= invoice.currency;
            totalOutstanding += invoice.amountDue ?? 0;
            invoiceCount += 1;
          }

          if (!res.hasMore) {
            return currency
              ? { totalOutstanding, invoiceCount, currency }
              : { totalOutstanding, invoiceCount };
          }
          if (newInvoiceCount === 0) return null;
        }

        // Never expose a partial total when the server keeps reporting more.
        return null;
      } catch {
        return null;
      }
    },
    enabled: !!customerId,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
