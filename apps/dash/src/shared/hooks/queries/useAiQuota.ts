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

type RawAiUsage = Partial<AiUsage> & {
  plan_slug?: string;
  allowance_tokens?: string | number;
  used_tokens?: string | number;
  remaining_tokens?: string | number;
  overage_tokens?: string | number;
  allow_overage?: boolean;
  period_end?: string;
};

type AiQuotaSettingsResponse = {
  settings?: { mode?: AiQuotaMode };
};

function finiteNumber(value: unknown): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function normalizeAiUsage(raw: RawAiUsage | null | undefined): AiUsage {
  return {
    planSlug: raw?.planSlug ?? raw?.plan_slug ?? 'free',
    mode: raw?.mode === 'per_person' ? 'per_person' : 'pool',
    allowanceTokens: finiteNumber(raw?.allowanceTokens ?? raw?.allowance_tokens),
    usedTokens: finiteNumber(raw?.usedTokens ?? raw?.used_tokens),
    remainingTokens: finiteNumber(raw?.remainingTokens ?? raw?.remaining_tokens),
    overageTokens: finiteNumber(raw?.overageTokens ?? raw?.overage_tokens),
    allowOverage: Boolean(raw?.allowOverage ?? raw?.allow_overage),
    blocked: Boolean(raw?.blocked),
    periodEnd: raw?.periodEnd ?? raw?.period_end ?? '',
  };
}

export const useAiUsage = () =>
  useQuery({
    queryKey: ['ai-usage'],
    queryFn: async (): Promise<AiUsage> => {
      const res = await unwrap(safe.get<{ usage?: RawAiUsage }>('/hub/ai/v1/usage'));
      return normalizeAiUsage(res?.usage);
    },
    staleTime: 30_000,
  });

export const useAiQuotaMode = () =>
  useQuery({
    queryKey: ['ai-quota-settings'],
    queryFn: async (): Promise<AiQuotaMode> => {
      const res = await unwrap(
        safe.get<{ settings: { mode: AiQuotaMode } }>('/hub/ai/v1/quota-settings')
      );
      return res?.settings?.mode ?? 'pool';
    },
    staleTime: 30_000,
  });

export const useUpdateAiQuotaMode = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (mode: AiQuotaMode) =>
      unwrap<AiQuotaSettingsResponse>(safe.put('/hub/ai/v1/quota-settings', { mode })),
    onSuccess: (res, requestedMode) => {
      const returnedMode = res?.settings?.mode;
      const nextMode =
        returnedMode === 'pool' || returnedMode === 'per_person' ? returnedMode : requestedMode;
      qc.setQueryData<AiQuotaMode>(['ai-quota-settings'], nextMode);
      qc.invalidateQueries({ queryKey: ['ai-usage'] });
      toast.success('AI allocation updated');
    },
    onError: () => toast.error('Failed to update AI allocation'),
  });
};
