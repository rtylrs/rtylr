import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import { campaignSchema, campaignStatsSchema } from '../../schemas/campaigns';
import type { Campaign, CampaignStats } from '../../schemas/campaigns';

interface UpdateCampaignPayload {
  id: number;
  data: Partial<Campaign>;
}

interface CampaignStatusPayload {
  id: number;
  status: string;
}

export const useListCampaigns = (): UseQueryResult<Campaign[]> =>
  useQuery<Campaign[]>({
    queryKey: ['crm-campaigns'],
    queryFn: async (): Promise<Campaign[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/campaigns'));
      return z.array(campaignSchema).parse(unwrapList(res, 'campaigns'));
    },
  });

export const useGetCampaign = (id: number | undefined): UseQueryResult<Campaign> =>
  useQuery<Campaign>({
    queryKey: ['crm-campaign', id],
    queryFn: async (): Promise<Campaign> => {
      const res = await unwrap(safe.get<unknown>('/crm/campaigns/detail', { params: { id } }));
      return campaignSchema.parse(res);
    },
    enabled: !!id,
  });

export const useCampaignStats = (id: number | undefined): UseQueryResult<CampaignStats> =>
  useQuery<CampaignStats>({
    queryKey: ['crm-campaign', id, 'stats'],
    queryFn: async (): Promise<CampaignStats> => {
      const res = await unwrap(safe.get<unknown>('/crm/campaigns/stats', { params: { id } }));
      return campaignStatsSchema.parse(res);
    },
    enabled: !!id,
  });

export const useCreateCampaign = (): UseMutationResult<Campaign, unknown, Partial<Campaign>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Campaign, unknown, Partial<Campaign>>({
    mutationFn: async (data: Partial<Campaign>): Promise<Campaign> => {
      const res = await unwrap(safe.post<Campaign>('/crm/campaigns', data));
      return res as Campaign;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      toast.success(t('crm.campaigns.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateCampaign = (): UseMutationResult<
  Campaign,
  unknown,
  UpdateCampaignPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Campaign, unknown, UpdateCampaignPayload>({
    mutationFn: async (payload: UpdateCampaignPayload): Promise<Campaign> => {
      const res = await unwrap(
        safe.post<Campaign>('/crm/campaigns', { id: payload.id, ...payload.data })
      );
      return res as Campaign;
    },
    onSuccess: (_data: Campaign, variables: UpdateCampaignPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      void qc.invalidateQueries({ queryKey: ['crm-campaign', variables.id] });
      toast.success(t('crm.campaigns.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteCampaign = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.delete('/crm/campaigns', { data: { id } })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      toast.success(t('crm.campaigns.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateCampaignStatus = (): UseMutationResult<
  unknown,
  unknown,
  CampaignStatusPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, CampaignStatusPayload>({
    mutationFn: (payload: CampaignStatusPayload): Promise<unknown> =>
      unwrap(safe.post('/crm/campaigns/status', payload)),
    onSuccess: (_data: unknown, variables: CampaignStatusPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      void qc.invalidateQueries({ queryKey: ['crm-campaign', variables.id] });
      toast.success(t('crm.campaigns.statusUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Campaign, CampaignStats } from '../../schemas/campaigns';
