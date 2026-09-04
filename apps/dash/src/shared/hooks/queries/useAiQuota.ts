import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { aiClient } from '../../lib/clients';

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

type RawAiUsage = {
  planSlug?: string;
  mode?: string;
  allowanceTokens?: string | number;
  usedTokens?: string | number;
  remainingTokens?: string | number;
  overageTokens?: string | number;
  allowOverage?: boolean;
  blocked?: boolean;
  periodEnd?: string;
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
      const res = await aiClient.getAiUsage({});
      return normalizeAiUsage(res?.usage);
    },
    staleTime: 30_000,
  });

export const useAiQuotaMode = () =>
  useQuery({
    queryKey: ['ai-quota-settings'],
    queryFn: async (): Promise<AiQuotaMode> => {
      const res = await aiClient.getAiQuotaSettings({});
      return (res.settings?.mode as AiQuotaMode | undefined) ?? 'pool';
    },
    staleTime: 30_000,
  });

import i18next from 'i18next';

export const useUpdateAiQuotaMode = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (mode: AiQuotaMode) => aiClient.updateAiQuotaSettings({ mode }),
    onSuccess: (res, requestedMode) => {
      const returnedMode = res?.settings?.mode;
      const nextMode =
        returnedMode === 'pool' || returnedMode === 'per_person' ? returnedMode : requestedMode;
      qc.setQueryData<AiQuotaMode>(['ai-quota-settings'], nextMode);
      qc.invalidateQueries({ queryKey: ['ai-usage'] });
      toast.success(i18next.t('common.toasts.aiAllocationUpdated', 'AI allocation updated'));
    },
    onError: () =>
      toast.error(
        i18next.t('common.toasts.failedUpdateAiAllocation', 'Failed to update AI allocation')
      ),
  });
};
