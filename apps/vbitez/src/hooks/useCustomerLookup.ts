import { useEffect, useState } from 'react';
import { posService } from '../services/posService';
import type { Customer } from '../types/pos';

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
        const res = await posService.searchCustomers(query.trim());
        if (cancelled) return;
        const raw = res?.data?.customers ?? res?.data ?? [];
        const mapped: Customer[] = (Array.isArray(raw) ? raw : []).map((c: any) => ({
          id: String(c.id ?? Math.random().toString(36).slice(2)),
          name: c.name ?? 'Customer',
          phone: c.phone,
          email: c.email,
          loyaltyPoints: c.loyaltyPoints,
          lastVisit: c.lastVisit,
        }));
        setResults(mapped);
      } catch (e: any) {
        if (!cancelled) {
          setError(e?.response?.data?.message ?? e?.message ?? 'Search failed');
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
