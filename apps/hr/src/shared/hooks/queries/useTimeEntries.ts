import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListTimeEntriesRequest,
  ListTimeEntriesResponse,
  CreateTimeEntryRequest,
  CreateTimeEntryResponse,
  TimeEntryPayload,
} from '@voxire/contracts/hr/v1/time_entries';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useListTimeEntries = (params: ListTimeEntriesRequest) => {
  return useQuery<ListTimeEntriesResponse>({
    queryKey: ['time-entries', 'filter', params],
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
      const res = await unwrap(
        safe.put<{ entry: TimeEntryPayload }>(`/hr/time-entries/${data.id}`, data)
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
