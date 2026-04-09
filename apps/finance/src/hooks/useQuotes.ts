import { useQuery } from '@tanstack/react-query';
import { MOCK_QUOTES, Quote } from '../lib/mockData';

export function useQuotes() {
  return useQuery<Quote[]>({
    queryKey: ['quotes'],
    queryFn: async () => MOCK_QUOTES,
  });
}
