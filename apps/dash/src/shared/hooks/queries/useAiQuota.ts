import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';

export type AiQuotaMode = 'pool' | 'per_person';

export interface AiUsage {
  planSlug: string;
  mode: AiQuotaMode;
  allowanceTokens: number;
  usedTokens: number;
  remainingTokens: number;
  overageTokens: number;
  allowOverage: boolean;
  blocked: boolean;
  periodEnd: string;
}

export const useAiUsage = () =>
  useQuery({
    queryKey: ['ai-usage'],
    queryFn: async (): Promise<AiUsage> => {
      const res = await unwrap(safe.get<{ usage: AiUsage }>('/api/hub/ai/v1/usage'));
      return (res as { usage: AiUsage }).usage;
    },
    staleTime: 30_000,
  });

export const useAiQuotaMode = () =>
  useQuery({
    queryKey: ['ai-quota-settings'],
    queryFn: async (): Promise<AiQuotaMode> => {
      const res = await unwrap(safe.get<{ settings: { mode: AiQuotaMode } }>('/api/hub/ai/v1/quota-settings'));
      return res?.settings?.mode ?? 'pool';
    },
    staleTime: 30_000,
  });

export const useUpdateAiQuotaMode = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (mode: AiQuotaMode) => unwrap(safe.put('/api/hub/ai/v1/quota-settings', { mode })),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['ai-quota-settings'] });
      qc.invalidateQueries({ queryKey: ['ai-usage'] });
      toast.success('AI allocation updated');
    },
    onError: () => toast.error('Failed to update AI allocation'),
  });
};
