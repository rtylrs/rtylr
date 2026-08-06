import { createOneKitClient } from '@voxire/api';
import { CustomerServiceClient } from '@voxire/voxie/pos/customer/v1/client';
import { useEffect, useState } from 'react';
import type { Customer } from '../../shared/types/pos';
import { mapCustomers } from './customerLookupWorkflow';

const customerClient = createOneKitClient(CustomerServiceClient, 'pos');

export function useCustomerLookup(initialQuery = '') {
  const [query, setQuery] = useState(initialQuery);
  const [results, setResults] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    if (!query.trim()) {
      setResults([]);
      return;
    }
    const handle = window.setTimeout(async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await customerClient.searchCustomers({ query: query.trim() });
        if (cancelled) return;
        setResults(mapCustomers(res.customers ?? []));
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        if (!cancelled) {
          setError(err?.response?.data?.message ?? err?.message ?? 'Search failed');
          setResults([]);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }, 300);
    return () => {
      cancelled = true;
      window.clearTimeout(handle);
    };
  }, [query]);

  return { query, setQuery, results, loading, error };
}
