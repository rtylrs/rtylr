import { unwrap } from '@voxire/api';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/api/posService';
import type { Customer } from '../../shared/types/pos';
import { mapCustomers } from './customerLookupWorkflow';

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
        const res = await unwrap(posService.searchCustomers(query.trim()));
        if (cancelled) return;
        const raw = res?.customers ?? res?.data ?? [];
        setResults(mapCustomers(raw));
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
