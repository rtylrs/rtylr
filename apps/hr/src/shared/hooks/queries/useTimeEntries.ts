import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListTimeEntriesRequest,
  ListTimeEntriesResponse,
} from '@voxire/voxie/hr/time_entry/v1/types';
import type {
  CreateTimeEntryRequest,
  CreateTimeEntryResponse,
} from '@voxire/voxie/hr/time_entry/v1/types';
import type { TimeEntryPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useListTimeEntries = (
  params: ListTimeEntriesRequest,
  options?: { enabled?: boolean }
) => {
  return useQuery<ListTimeEntriesResponse>({
    queryKey: ['time-entries', 'filter', params],
    enabled: options?.enabled ?? true,
    queryFn: async (): Promise<ListTimeEntriesResponse> => {
      const res = await unwrap(safe.get<ListTimeEntriesResponse>('/hr/time-entries', { params }));
      return res as ListTimeEntriesResponse;
    },
  });
};

export const useCreateTimeEntry = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateTimeEntryRequest): Promise<CreateTimeEntryResponse> => {
      const res = await unwrap(safe.post<CreateTimeEntryResponse>('/hr/time-entries', data));
      return res as CreateTimeEntryResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['time-entries'] });
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      toast.success(t('timeLogging.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateTimeEntry = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: Partial<CreateTimeEntryRequest> & { id: number }
    ): Promise<TimeEntryPayload> => {
      const { id, ...body } = data;
      const res = await unwrap(
        safe.put<{ entry: TimeEntryPayload }>(`/hr/time-entries/${id}`, body)
      );
      return (res as { entry: TimeEntryPayload }).entry;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['time-entries'] });
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      toast.success(t('timeLogging.updated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

interface FlowImportResponse {
  imported: number;
  skipped: number;
}

export const useImportFromFlow = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (): Promise<FlowImportResponse> => {
      const res = await unwrap(safe.post<FlowImportResponse>('/hr/time-entries/import-flow', {}));
      return res as FlowImportResponse;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['time-entries'] });
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });

      if (data.imported === 0 && data.skipped === 0) {
        toast.info(t('timeLogging.flowImport.noEntries'));
      } else if (data.skipped > 0) {
        toast.success(
          t('timeLogging.flowImport.successWithSkipped', {
            imported: data.imported,
            skipped: data.skipped,
          })
        );
      } else {
        toast.success(t('timeLogging.flowImport.success', { imported: data.imported }));
      }
    },
    onError: () => {
      toast.error(t('timeLogging.flowImport.error'));
    },
  });
};

export const useDeleteTimeEntry = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<void> => {
      await unwrap(safe.delete(`/hr/time-entries/${id}`));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['time-entries'] });
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      toast.success(t('timeLogging.deleted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
