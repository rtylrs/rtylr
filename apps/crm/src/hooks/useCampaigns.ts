import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Campaign } from '../services/crmApi';

export function useCampaigns() {
  return useQuery({
    queryKey: ['campaigns'],
    queryFn: () => crmApi.getCampaigns(),
  });
}

export function useCampaignById(id: number | undefined) {
  return useQuery({
    queryKey: ['campaign', id],
    queryFn: () => crmApi.getCampaignById(id!),
    enabled: !!id,
  });
}

export function useCampaignStats(id: number | undefined) {
  return useQuery({
    queryKey: ['campaign-stats', id],
    queryFn: () => crmApi.getCampaignStats(id!),
    enabled: !!id,
  });
}

export function useCreateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Campaign>) => crmApi.createCampaign(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}

export function useUpdateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Campaign> }) => crmApi.updateCampaign(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}

export function useUpdateCampaignStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) => crmApi.updateCampaignStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['campaigns'] }),
  });
}
