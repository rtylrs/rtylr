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
import {
  loyaltyConfigSchema,
  loyaltyMemberSchema,
  loyaltyRewardSchema,
  loyaltyTierSchema,
  loyaltyTransactionSchema,
} from '../../schemas/loyalty';
import type {
  LoyaltyConfig,
  LoyaltyMember,
  LoyaltyReward,
  LoyaltyTier,
  LoyaltyTransaction,
} from '../../schemas/loyalty';

// ── Types ────────────────────────────────────────────────────────────────────

interface UpsertTierPayload {
  id?: number;
  data: Partial<LoyaltyTier>;
}

interface UpsertRewardPayload {
  id?: number;
  data: Partial<LoyaltyReward>;
}

interface AwardRedeemPayload {
  customerId: number;
  points: number;
  reason?: string;
}

function toUpdateLoyaltyConfigRequest(config: Partial<LoyaltyConfig>): Record<string, unknown> {
  return {
    isEnabled: config.programEnabled ?? false,
    pointsPerCurrencyUnit: config.pointsPerDollar ?? 1,
    currencyPerPoint: config.redemptionRate ?? 0.01,
    bronzeMin: config.bronzeMin ?? 0,
    silverMin: config.silverMin ?? 500,
    goldMin: config.goldMin ?? 1500,
    platinumMin: config.platinumMin ?? 5000,
    bronzeDiscountPct: config.bronzeDiscountPct ?? 0,
    silverDiscountPct: config.silverDiscountPct ?? 5,
    goldDiscountPct: config.goldDiscountPct ?? 10,
    platinumDiscountPct: config.platinumDiscountPct ?? 15,
    pointsExpiryDays: config.pointsExpiry ?? 365,
  };
}

function toLoyaltyTierRequest(tier: Partial<LoyaltyTier>): Record<string, unknown> {
  return {
    name: tier.name ?? '',
    minPoints: tier.threshold ?? 0,
    discountPct: tier.discountPct ?? tier.multiplier ?? 0,
    perks: JSON.stringify(tier.perks ?? []),
    sortOrder: tier.sortOrder ?? 0,
  };
}

// ── Queries ──────────────────────────────────────────────────────────────────

export const useLoyaltyConfig = (): UseQueryResult<LoyaltyConfig> =>
  useQuery<LoyaltyConfig>({
    queryKey: ['crm-loyalty-config'],
    queryFn: async (): Promise<LoyaltyConfig> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/config'));
      const envelope = (res ?? {}) as { config?: unknown };
      return loyaltyConfigSchema.parse(envelope.config ?? res);
    },
  });

export const useLoyaltyTiers = (): UseQueryResult<LoyaltyTier[]> =>
  useQuery<LoyaltyTier[]>({
    queryKey: ['crm-loyalty-tiers'],
    queryFn: async (): Promise<LoyaltyTier[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/tiers'));
      return z.array(loyaltyTierSchema).parse(unwrapList(res, 'tiers'));
    },
  });

export const useLoyaltyMembers = (limit = 100, offset = 0): UseQueryResult<LoyaltyMember[]> =>
  useQuery<LoyaltyMember[]>({
    queryKey: ['crm-loyalty-members', { limit, offset }],
    queryFn: async (): Promise<LoyaltyMember[]> => {
      const res = await unwrap(
        safe.get<unknown>('/crm/loyalty/members', { params: { limit, offset } })
      );
      return z.array(loyaltyMemberSchema).parse(unwrapList(res, 'members'));
    },
  });

export const useLoyaltyRewards = (): UseQueryResult<LoyaltyReward[]> =>
  useQuery<LoyaltyReward[]>({
    queryKey: ['crm-loyalty-rewards'],
    queryFn: async (): Promise<LoyaltyReward[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/rewards'));
      return z.array(loyaltyRewardSchema).parse(unwrapList(res, 'rewards'));
    },
  });

export const useLoyaltyTransactions = (customerId?: number): UseQueryResult<LoyaltyTransaction[]> =>
  useQuery<LoyaltyTransaction[]>({
    queryKey: ['crm-loyalty-transactions', customerId],
    queryFn: async (): Promise<LoyaltyTransaction[]> => {
      if (!customerId) return [];
      const res = await unwrap(
        safe.get<unknown>('/crm/loyalty/transactions', {
          params: { customer_id: customerId, page: 1, page_size: 100 },
        })
      );
      return z.array(loyaltyTransactionSchema).parse(unwrapList(res, 'transactions'));
    },
    enabled: typeof customerId === 'number' && customerId > 0,
  });

// ── Mutations ────────────────────────────────────────────────────────────────

export const useUpdateLoyaltyConfig = (): UseMutationResult<
  LoyaltyConfig,
  unknown,
  Partial<LoyaltyConfig>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<LoyaltyConfig, unknown, Partial<LoyaltyConfig>>({
    mutationFn: async (config: Partial<LoyaltyConfig>): Promise<LoyaltyConfig> => {
      const current = qc.getQueryData<LoyaltyConfig>(['crm-loyalty-config']) ?? {};
      const request = toUpdateLoyaltyConfigRequest({ ...current, ...config });
      await unwrap(safe.put('/crm/loyalty/config', request));
      return loyaltyConfigSchema.parse({ ...current, ...config });
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-config'] });
      toast.success(t('crm.loyalty.configUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpsertTier = (): UseMutationResult<unknown, unknown, UpsertTierPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, UpsertTierPayload>({
    mutationFn: (payload: UpsertTierPayload): Promise<unknown> => {
      const current = qc
        .getQueryData<LoyaltyTier[]>(['crm-loyalty-tiers'])
        ?.find((tier) => tier.id === payload.id);
      const request = { id: payload.id, ...toLoyaltyTierRequest({ ...current, ...payload.data }) };
      return payload.id
        ? unwrap(safe.put('/crm/loyalty/tiers', request))
        : unwrap(safe.post('/crm/loyalty/tiers', request));
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-tiers'] });
      toast.success(t('crm.loyalty.tierSaved'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpsertReward = (): UseMutationResult<unknown, unknown, UpsertRewardPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, UpsertRewardPayload>({
    mutationFn: (payload: UpsertRewardPayload): Promise<unknown> =>
      unwrap(
        safe.post(
          '/crm/loyalty/rewards/upsert',
          payload.id ? { id: payload.id, ...payload.data } : payload.data
        )
      ),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-rewards'] });
      toast.success(t('crm.loyalty.rewardSaved'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useAwardPoints = (): UseMutationResult<unknown, unknown, AwardRedeemPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, AwardRedeemPayload>({
    mutationFn: (payload: AwardRedeemPayload): Promise<unknown> => {
      const request = {
        customerId: payload.customerId,
        points: payload.points,
        description: payload.reason ?? '',
        referenceType: 'manual',
      };
      return unwrap(safe.post('/crm/loyalty/award', request));
    },
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.customerId] });
      toast.success(t('crm.loyalty.pointsAwarded'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useRedeemPoints = (): UseMutationResult<unknown, unknown, AwardRedeemPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, AwardRedeemPayload>({
    mutationFn: (payload: AwardRedeemPayload): Promise<unknown> => {
      const request = {
        customerId: payload.customerId,
        points: payload.points,
        description: payload.reason ?? '',
        referenceType: 'manual',
      };
      return unwrap(safe.post('/crm/loyalty/redeem', request));
    },
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.customerId] });
      toast.success(t('crm.loyalty.pointsRedeemed'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type {
  LoyaltyConfig,
  LoyaltyMember,
  LoyaltyReward,
  LoyaltyTier,
  LoyaltyTransaction,
} from '../../schemas/loyalty';
