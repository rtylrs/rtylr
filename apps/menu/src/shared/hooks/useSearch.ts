import { useState, useMemo } from 'react';
import { filterByField, sortByRelevance } from '../utils/searchUtils';

interface UseSearchReturn<T> {
  query: string;
  setQuery: (value: string) => void;
  results: T[];
  clear: () => void;
}

/**
 * Filter and rank a list of items by a text query.
 * Items are filtered using case-insensitive substring matching and ranked so
 * that items whose field starts with the query appear first.
 */
export function useSearch<T>(items: T[], getField: (item: T) => string): UseSearchReturn<T> {
  const [query, setQuery] = useState('');

  const results = useMemo(() => {
    const filtered = filterByField(items, getField, query);
    return sortByRelevance(filtered, getField, query);
  }, [items, getField, query]);

  return {
    query,
    setQuery,
    results,
    clear: () => setQuery(''),
  };
}
