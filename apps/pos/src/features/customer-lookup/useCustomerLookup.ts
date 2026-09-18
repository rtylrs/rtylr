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
    const trimmed = query.trim();
    // Don't hit the search API for a single character.
    if (trimmed.length < 2) {
      setResults([]);
      setLoading(false);
      return;
    }
    const handle = window.setTimeout(async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await customerClient.searchCustomers({ query: trimmed });
        if (cancelled) return;
        setResults(mapCustomers(res.customers ?? []));
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error && e.message ? e.message : 'Search failed');
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
