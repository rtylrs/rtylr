import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { CustomerServiceClient } from '@voxire/voxie/support/customer/v1/client';

// Support is a foreign domain: read-only lookup so the CRM customer profile
// can show support engagement. Client app id stays 'crm'.
const supportCustomerClient = createOneKitClient(CustomerServiceClient, 'crm');

export interface SupportEngagement {
  totalTickets: number;
  lastTicketAt?: string;
}

export const SUPPORT_ENGAGEMENT_KEY = (email: string) => ['support', 'engagement', email] as const;

/**
 * Support ticket count and last-ticket date for one customer, matched by
 * email (#842 / rtylr#861). Email matching is the only mechanism today; the
 * verified customer_id link (voxie#51) exists in the schema but no lookup-by-
 * CRM-customer RPC has been contracted yet. Once it is, swap the queryFn to
 * use it instead of scanning.
 *
 * Degrade: a tenant without the support module — or a failed read — yields
 * null and the panel hides; never an empty-zero claim about ticket history.
 */
export function useSupportEngagement(email?: string | null) {
  return useQuery<SupportEngagement | null>({
    queryKey: SUPPORT_ENGAGEMENT_KEY(email ?? ''),
    queryFn: async (): Promise<SupportEngagement | null> => {
      if (!email) return null;
      try {
        const res = await supportCustomerClient.listCustomers({ page: 1, pageSize: 500 });
        if (res === null || res === undefined) return null;
        const match = (res.customers ?? []).find(
          (c) => c.email?.toLowerCase() === email.toLowerCase()
        );
        if (!match) return null;
        return {
          totalTickets: match.totalTickets ?? 0,
          lastTicketAt: match.lastTicketAt,
        };
      } catch {
        return null;
      }
    },
    enabled: !!email,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
