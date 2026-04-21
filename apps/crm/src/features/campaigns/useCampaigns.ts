import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, Campaign } from '../../shared/services/crmApi';
import { campaignSchema, campaignStatsSchema } from '../../shared/schemas/campaigns';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useCampaigns() {
  return useQuery({
    queryKey: ['campaigns'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCampaigns());
      return z.array(campaignSchema).parse(unwrapList(res, 'campaigns'));
    },
  });
}

export function useCampaignById(id: number | undefined) {
  return useQuery({
    queryKey: ['campaign', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCampaignById(id!));
      return campaignSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useCampaignStats(id: number | undefined) {
  return useQuery({
    queryKey: ['campaign-stats', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCampaignStats(id!));
      return campaignStatsSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useCreateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Campaign>) => unwrap(crmApi.createCampaign(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}

export function useUpdateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Campaign> }) =>
      unwrap(crmApi.updateCampaign(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}

export function useUpdateCampaignStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      unwrap(crmApi.updateCampaignStatus(id, status)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}

export function useDeleteCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(crmApi.deleteCampaign(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}
