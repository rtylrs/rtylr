import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';
import { parseCrmDate } from '../../lib/dateUtils';
import { unwrapList } from '../../lib/unwrapList';
import { campaignSchema, campaignStatsSchema } from '../../schemas/campaigns';
import type { Campaign, CampaignStats } from '../../schemas/campaigns';

interface UpdateCampaignPayload {
  id: number;
  data: Partial<Campaign>;
}

interface CampaignFormData extends Partial<Campaign> {
  channel?: string;
  content?: string;
  description?: string;
  segmentId?: number | string | null;
  scheduledFor?: string;
}

interface CampaignStatusPayload {
  id: number;
  status: string;
}

interface CampaignRequest {
  name: string;
  type: string;
  audienceSegmentId?: number;
  subject: string;
  body: string;
  scheduledAt?: string;
}

export interface CampaignMutationResponse {
  campaignId?: number;
  campaign_id?: number;
  message?: string;
}

function normalizeScheduledAt(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  if (!trimmed) return undefined;
  const date = parseCrmDate(trimmed);
  if (!Number.isFinite(date.getTime())) return undefined;
  return date.toISOString().replace(/\.\d{3}Z$/, 'Z');
}

export function toCampaignRequest(data: CampaignFormData): CampaignRequest {
  const name = data.name?.trim() ?? '';
  const subject = data.subject?.trim() || name;
  const segmentId = data.audienceSegmentId ?? data.segmentId;
  const parsedSegmentId = Number(segmentId);
  return {
    name,
    type: data.type ?? data.channel ?? 'email',
    audienceSegmentId:
      Number.isInteger(parsedSegmentId) && parsedSegmentId > 0 ? parsedSegmentId : undefined,
    subject,
    body: data.body ?? data.content ?? '',
    scheduledAt: normalizeScheduledAt(data.scheduledAt ?? data.scheduledFor),
  };
}

function fromCampaignPayload(campaign: unknown): Campaign {
  return campaignSchema.parse(campaign);
}

function responseRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' ? (value as Record<string, unknown>) : {};
}

function campaignIdFromResponse(value: unknown): number | undefined {
  const raw = responseRecord(value);
  const candidate = raw.campaignId ?? raw.campaign_id;
  const id = Number(candidate);
  return Number.isFinite(id) && id > 0 ? id : undefined;
}

function messageFromResponse(value: unknown): string | undefined {
  const message = responseRecord(value).message;
  return typeof message === 'string' ? message : undefined;
}

export const useListCampaigns = (limit = 100, offset = 0): UseQueryResult<Campaign[]> =>
  useQuery<Campaign[]>({
    queryKey: ['crm-campaigns', { limit, offset }],
    queryFn: async (): Promise<Campaign[]> => {
      const res = await unwrap(
        safe.get<unknown>('/crm/campaigns', {
          params: {
            limit: Math.max(1, Math.floor(limit)),
            offset: Math.max(0, Math.floor(offset)),
          },
        })
      );
      return unwrapList<unknown>(res, 'campaigns').map(fromCampaignPayload);
    },
  });

export const useGetCampaign = (id: number | undefined): UseQueryResult<Campaign> =>
  useQuery<Campaign>({
    queryKey: ['crm-campaign', id],
    queryFn: async (): Promise<Campaign> => {
      if (id == null || id <= 0) throw new Error('Invalid campaign id');
      const res = await unwrap(safe.get<unknown>(`/crm/campaigns/${id}`));
      const campaign = responseRecord(res).campaign ?? res;
      return fromCampaignPayload(campaign);
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

export const useCampaignStats = (id: number | undefined): UseQueryResult<CampaignStats> =>
  useQuery<CampaignStats>({
    queryKey: ['crm-campaign', id, 'stats'],
    queryFn: async (): Promise<CampaignStats> => {
      if (id == null || id <= 0) throw new Error('Invalid campaign id');
      const res = await unwrap(
        safe.get<unknown>('/crm/campaigns/stats', { params: { campaignId: id } })
      );
      const stats = responseRecord(res).stats ?? res;
      return campaignStatsSchema.parse(stats);
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
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
      const res = await unwrap(safe.post<unknown>('/crm/campaigns', toCampaignRequest(data)));
      return { campaignId: campaignIdFromResponse(res), message: messageFromResponse(res) };
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      toast.success(t('crm.campaigns.created'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
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
      const cached = qc.getQueryData<Campaign>(['crm-campaign', payload.id]);
      const res = await unwrap(
        safe.put<unknown>(
          `/crm/campaigns/${payload.id}`,
          toCampaignRequest({ ...cached, ...payload.data })
        )
      );
      return { message: messageFromResponse(res) };
    },
    onSuccess: (_data: CampaignMutationResponse, variables: UpdateCampaignPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      void qc.invalidateQueries({ queryKey: ['crm-campaign', variables.id] });
      toast.success(t('crm.campaigns.updated'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
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
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
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
      unwrap(safe.put(`/crm/campaigns/${payload.id}/status`, { status: payload.status })),
    onSuccess: (_data: unknown, variables: CampaignStatusPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-campaigns'] });
      void qc.invalidateQueries({ queryKey: ['crm-campaign', variables.id] });
      toast.success(t('crm.campaigns.statusUpdated'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export type { Campaign, CampaignStats } from '../../schemas/campaigns';
