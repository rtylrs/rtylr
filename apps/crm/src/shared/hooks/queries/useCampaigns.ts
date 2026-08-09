import { unwrap } from '@voxire/api';
import type { CreateCampaignRequest } from '@voxire/voxie/crm/campaign/v1/types';
import type { UpdateCampaignRequest } from '@voxire/voxie/crm/campaign/v1/types';
import type { UpdateCampaignStatusRequest } from '@voxire/voxie/crm/campaign/v1/types';
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
      // Backend campaigns module registers GET /crm/campaigns/{id} and wraps
      // the result in a `{ campaign: ... }` envelope.
      const res = await unwrap(safe.get<unknown>(`/crm/campaigns/${id}`));
      const envelope = (res ?? {}) as { campaign?: unknown };
      return campaignSchema.parse(envelope.campaign ?? res);
    },
    enabled: !!id,
  });

export const useCampaignStats = (id: number | undefined): UseQueryResult<CampaignStats> =>
  useQuery<CampaignStats>({
    queryKey: ['crm-campaign', id, 'stats'],
    queryFn: async (): Promise<CampaignStats> => {
      // The stats handler reads the `campaign_id` query parameter. Query params
      // are sent verbatim (the api interceptor only camel→snake-cases the body),
      // so the snake_case key must be passed literally here.
      const res = await unwrap(
        safe.get<unknown>('/crm/campaigns/stats', { params: { campaign_id: id } })
      );
      return campaignStatsSchema.parse(res);
    },
    enabled: !!id,
  });

export const useCreateCampaign = (): UseMutationResult<Campaign, unknown, Partial<Campaign>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Campaign, unknown, Partial<Campaign>>({
    mutationFn: async (data: Partial<Campaign>): Promise<Campaign> => {
      const res = await unwrap(
        safe.post<Campaign>('/crm/campaigns', data as CreateCampaignRequest)
      );
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
        safe.put<Campaign>(`/crm/campaigns/${payload.id}`, {
          id: String(payload.id),
          ...payload.data,
        } as UpdateCampaignRequest)
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
    mutationFn: (id: number): Promise<unknown> => unwrap(safe.delete(`/crm/campaigns/${id}`)),
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
    mutationFn: (payload: CampaignStatusPayload): Promise<unknown> => {
      return unwrap(
        safe.put(`/crm/campaigns/${payload.id}/status`, {
          id: String(payload.id),
          status: payload.status,
        } as UpdateCampaignStatusRequest)
      );
    },
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
