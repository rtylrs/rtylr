import { useQuery } from '@tanstack/react-query';
import { MOCK_ACCOUNTS, MOCK_JOURNAL, Account, JournalEntry } from '../lib/mockData';

export function useChartOfAccounts() {
  return useQuery<Account[]>({
    queryKey: ['chart-of-accounts'],
    queryFn: async () => MOCK_ACCOUNTS,
  });
}

export function useJournal() {
  return useQuery<JournalEntry[]>({
    queryKey: ['journal'],
    queryFn: async () => MOCK_JOURNAL,
  });
}
