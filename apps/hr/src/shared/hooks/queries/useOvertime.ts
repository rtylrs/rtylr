import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { OvertimeRequest, OvertimeConfig } from '../../schemas';

interface ListOvertimeRequestsParams {
  status?: string;
  staffId?: number;
}

interface ListOvertimeRequestsResponse {
  requests: OvertimeRequest[];
}

interface CreateOvertimeRequestPayload {
  date: string;
  plannedMinutes: number;
  reason: string;
}

interface ApproveOvertimePayload {
  id: number;
  status: string;
  notes?: string;
}

interface OvertimeConfigResponse {
  config: OvertimeConfig;
}

export const useListOvertimeRequests = (params?: ListOvertimeRequestsParams) => {
  return useQuery<OvertimeRequest[]>({
    queryKey: ['overtime-requests', params],
    queryFn: async (): Promise<OvertimeRequest[]> => {
      const res = await unwrap(
        safe.get<ListOvertimeRequestsResponse>('/hr/overtime/requests', { params })
      );
      return (res as ListOvertimeRequestsResponse).requests;
    },
  });
};

export const useCreateOvertimeRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateOvertimeRequestPayload): Promise<void> => {
      await unwrap(safe.post('/hr/overtime/requests', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveOvertime = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApproveOvertimePayload): Promise<void> => {
      await unwrap(safe.put(`/hr/overtime/requests/${data.id}/approve`, data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetOvertimeConfig = () => {
  return useQuery<OvertimeConfig>({
    queryKey: ['overtime-config'],
    queryFn: async (): Promise<OvertimeConfig> => {
      const res = await unwrap(safe.get<OvertimeConfigResponse>('/hr/overtime/config'));
      return (res as OvertimeConfigResponse).config;
    },
  });
};

export const useUpdateOvertimeConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: Partial<OvertimeConfig>): Promise<void> => {
      await unwrap(safe.put('/hr/overtime/config', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-config'] });
      toast.success(t('overtime.configUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
