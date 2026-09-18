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
import { hrTransport } from '../../lib/clients';
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
      const res = await unwrap(
        hrTransport.get<ListTimeEntriesResponse>('/hr/time-entries', { params })
      );
      return res as ListTimeEntriesResponse;
    },
  });
};

export const useCreateTimeEntry = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateTimeEntryRequest): Promise<CreateTimeEntryResponse> => {
      const res = await unwrap(hrTransport.post<CreateTimeEntryResponse>('/hr/time-entries', data));
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
        hrTransport.put<{ entry: TimeEntryPayload }>(`/hr/time-entries/${id}`, body)
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
      await unwrap(hrTransport.delete(`/hr/time-entries/${id}`));
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
