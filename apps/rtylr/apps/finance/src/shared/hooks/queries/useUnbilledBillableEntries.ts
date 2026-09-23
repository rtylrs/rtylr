import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { TimeEntryServiceClient } from '@voxire/voxie/hr/time_entry/v1/client';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// HR is a foreign domain here: read-only listing plus one idempotent
// write-back that closes the billable loop. Client app id stays 'finance'.
const timeEntryClient = createOneKitClient(TimeEntryServiceClient, 'finance');

export interface UnbilledBillableEntry {
  id: string;
  date?: string;
  description?: string;
  durationMinutes?: number;
  projectId?: string;
}

export const UNBILLED_ENTRIES_QUERY_KEY = (from: string, to: string) =>
  ['hr', 'time-entries', 'unbilled', { from, to }] as const;

const UNBILLED_ENTRIES_PAGE_SIZE = 200;

/**
 * Approved billable time entries with no invoice yet, for the range.
 *
 * Degrade (#836): a tenant without HR (or a failed read) yields null and the
 * section hides; it never renders an empty-zero claim about someone's hours.
 */
export function useUnbilledBillableEntries(from: string, to: string) {
  return useQuery<UnbilledBillableEntry[] | null>({
    queryKey: UNBILLED_ENTRIES_QUERY_KEY(from, to),
    queryFn: async (): Promise<UnbilledBillableEntry[] | null> => {
      try {
        const entries: UnbilledBillableEntry[] = [];
        let offset = 0;

        while (true) {
          const res = await timeEntryClient.listTimeEntries({
            isBillable: true,
            status: 'approved',
            fromDate: from,
            toDate: to,
            limit: UNBILLED_ENTRIES_PAGE_SIZE,
            offset,
          });
          if (res === null || res === undefined) return null;

          const page = res.entries ?? [];
          entries.push(
            ...page.map((e) => ({
              id: e.id ?? '',
              date: e.date,
              description: e.description,
              durationMinutes: e.durationMinutes,
              projectId: e.projectId,
            }))
          );
          offset += page.length;
          const total = res.total ?? entries.length;

          if (page.length === 0 || !Number.isFinite(total) || offset >= total) break;
        }

        return entries;
      } catch {
        return null;
      }
    },
    staleTime: 2 * 60 * 1000,
    retry: false,
  });
}

/**
 * Closes the loop after an invoice is raised for the selected entries:
 * stamps each entry with the invoice id so it stops appearing unbilled.
 * Idempotent server-side per entry (already-invoiced ids are skipped, never
 * overwritten), so a retry after a partial failure is safe.
 */
export function useMarkEntriesInvoiced() {
  const qc = useQueryClient();
  const { t } = useTranslation('finance');
  return useMutation({
    mutationFn: async ({ entryIds, invoiceId }: { entryIds: string[]; invoiceId: string }) => {
      const res = await timeEntryClient.markTimeEntriesInvoiced({
        body: { entryIds, invoiceId },
      });
      return res;
    },
    onSuccess: (res) => {
      void qc.invalidateQueries({ queryKey: ['hr', 'time-entries'] });
      toast.success(
        t('finance.unbilled.markedDone', {
          marked: res.markedCount ?? 0,
          skipped: res.skippedCount ?? 0,
        })
      );
    },
    onError: () => toast.error(t('errors.generic')),
  });
}
