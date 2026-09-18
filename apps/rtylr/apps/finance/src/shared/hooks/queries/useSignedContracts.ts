import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ContractServiceClient } from '@voxire/voxie/crm/sales/v1/client';

// CRM is a foreign domain: read-only listing of signed contracts so Finance
// can prefill an invoice. Client app id stays 'finance'.
const contractClient = createOneKitClient(ContractServiceClient, 'finance');

export interface SignedContract {
  id: string;
  plan?: string;
  mrr?: string;
  seats?: number;
  signedAt?: string;
  signedByName?: string;
}

export const SIGNED_CONTRACTS_QUERY_KEY = ['crm', 'sales-contracts', 'signed'] as const;

/**
 * Signed sales contracts for the invoice prefill (CRM->FIN-01, rtylr#839).
 * Only fully signed contracts are returned: unsigned drafts are not
 * billable.
 *
 * Degrade: a tenant without the CRM sales surface (or a failed read) yields
 * null and the prefill control hides; it never blocks manual invoicing.
 */
export function useSignedContracts() {
  return useQuery<SignedContract[] | null>({
    queryKey: SIGNED_CONTRACTS_QUERY_KEY,
    queryFn: async (): Promise<SignedContract[] | null> => {
      try {
        const res = await contractClient.listContracts({ signed: true, limit: 50, offset: 0 });
        if (res === null || res === undefined) return null;
        return (res.contracts ?? []).map((c) => ({
          id: c.id ?? '',
          plan: c.plan,
          mrr: c.mrr,
          seats: c.seats,
          signedAt: c.signedAt,
          signedByName: c.signedByName,
        }));
      } catch {
        return null;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
