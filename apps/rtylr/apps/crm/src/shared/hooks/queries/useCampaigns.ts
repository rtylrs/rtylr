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
import { crmApi } from '../../services/crmApi';
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

export interface CampaignMutationResponse {
  campaignId?: number;
  campaign_id?: number;
  message?: string;
}

export const useListCampaigns = (limit = 100, offset = 0): UseQueryResult<Campaign[]> =>
  useQuery<Campaign[]>({
    queryKey: ['crm-campaigns', { limit, offset }],
    queryFn: async (): Promise<Campaign[]> => {
      const res = await unwrap(crmApi.getCampaigns(limit, offset));
      return z.array(campaignSchema).parse(unwrapList(res, 'campaigns'));
    },
  });

export const useGetCampaign = (id: number | undefined): UseQueryResult<Campaign> =>
  useQuery<Campaign>({
    queryKey: ['crm-campaign', id],
    queryFn: async (): Promise<Campaign> => {
      const res = await unwrap(crmApi.getCampaignById(id ?? 0));
      const envelope = (res ?? {}) as { campaign?: unknown };
      return campaignSchema.parse(envelope.campaign ?? res);
    },
    enabled: !!id,
  });

export const useCampaignStats = (id: number | undefined): UseQueryResult<CampaignStats> =>
  useQuery<CampaignStats>({
    queryKey: ['crm-campaign', id, 'stats'],
    queryFn: async (): Promise<CampaignStats> => {
      const res = await unwrap(crmApi.getCampaignStats(id ?? 0));
      const envelope = (res ?? {}) as { stats?: unknown };
      return campaignStatsSchema.parse(envelope.stats ?? res);
    },
    enabled: !!id,
  });

export const useCreateCampaign = (): UseMutationResult<
  CampaignMutationResponse,
  unknown,
  Partial<Campaign>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<CampaignMutationResponse, unknown, Partial<Campaign>>({
    mutationFn: async (data: Partial<Campaign>): Promise<CampaignMutationResponse> => {
      const res = (await unwrap(
        crmApi.createCampaign(data as unknown as Partial<import('../../services/crmApi').Campaign>)
      )) as {
        campaignId?: number;
        campaign_id?: number;
        message?: string;
      };
      const campaignId = res.campaignId ?? res.campaign_id;
      return { campaignId, campaign_id: campaignId, message: res.message };
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
  CampaignMutationResponse,
  unknown,
  UpdateCampaignPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<CampaignMutationResponse, unknown, UpdateCampaignPayload>({
    mutationFn: async (payload: UpdateCampaignPayload): Promise<CampaignMutationResponse> => {
      const res = await unwrap(
        crmApi.updateCampaign(
          payload.id,
          payload.data as unknown as Partial<import('../../services/crmApi').Campaign>
        )
      );
      return res as CampaignMutationResponse;
    },
    onSuccess: (_data: CampaignMutationResponse, variables: UpdateCampaignPayload): void => {
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
    mutationFn: (id: number): Promise<unknown> => unwrap(crmApi.deleteCampaign(id)),
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
      return unwrap(crmApi.updateCampaignStatus(payload.id, payload.status));
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
