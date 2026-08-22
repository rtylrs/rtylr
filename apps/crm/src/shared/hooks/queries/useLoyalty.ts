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
import type { UpdateLoyaltyConfigRequest } from '@voxire/voxie/crm/loyalty/v1/types';
import { crmApi } from '../../services/crmApi';
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

function toUpdateLoyaltyConfigRequest(config: Partial<LoyaltyConfig>): UpdateLoyaltyConfigRequest {
  // UpdateLoyaltyConfigRequest marks every numeric field @gt(0), so zero
  // fallbacks would be rejected by the typed client before any network call;
  // the int64 threshold/expiry fields travel as strings on the wire.
  const positive = (value: number | undefined, fallback: number): number =>
    value !== undefined && Number.isFinite(value) && value > 0 ? value : fallback;
  return {
    isEnabled: config.programEnabled ?? false,
    pointsPerCurrencyUnit: positive(config.pointsPerDollar, 1),
    currencyPerPoint: positive(config.redemptionRate, 0.01),
    bronzeMin: String(positive(config.bronzeMin, 1)),
    silverMin: String(positive(config.silverMin, 500)),
    goldMin: String(positive(config.goldMin, 1500)),
    platinumMin: String(positive(config.platinumMin, 5000)),
    bronzeDiscountPct: positive(config.bronzeDiscountPct, 1),
    silverDiscountPct: positive(config.silverDiscountPct, 5),
    goldDiscountPct: positive(config.goldDiscountPct, 10),
    platinumDiscountPct: positive(config.platinumDiscountPct, 15),
    pointsExpiryDays: String(positive(config.pointsExpiry, 365)),
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
      const res = await unwrap(crmApi.getLoyaltyConfig());
      const envelope = (res ?? {}) as { config?: unknown };
      return loyaltyConfigSchema.parse(envelope.config ?? res);
    },
  });

export const useLoyaltyTiers = (): UseQueryResult<LoyaltyTier[]> =>
  useQuery<LoyaltyTier[]>({
    queryKey: ['crm-loyalty-tiers'],
    queryFn: async (): Promise<LoyaltyTier[]> => {
      const res = await unwrap(crmApi.getLoyaltyTiers());
      return z.array(loyaltyTierSchema).parse(unwrapList(res, 'tiers'));
    },
  });

export const useLoyaltyMembers = (limit = 100, offset = 0): UseQueryResult<LoyaltyMember[]> =>
  useQuery<LoyaltyMember[]>({
    queryKey: ['crm-loyalty-members', { limit, offset }],
    queryFn: async (): Promise<LoyaltyMember[]> => {
      const res = await unwrap(crmApi.getLoyaltyMembers(limit, offset));
      return z.array(loyaltyMemberSchema).parse(unwrapList(res, 'members'));
    },
  });

export const useLoyaltyRewards = (): UseQueryResult<LoyaltyReward[]> =>
  useQuery<LoyaltyReward[]>({
    queryKey: ['crm-loyalty-rewards'],
    queryFn: async (): Promise<LoyaltyReward[]> => {
      const res = await unwrap(crmApi.getLoyaltyRewards());
      return z.array(loyaltyRewardSchema).parse(unwrapList(res, 'rewards'));
    },
  });

// With no customerId this returns the business-wide recent activity feed
// (backend treats a missing/zero customer_id as "all customers"), which is
// what the Loyalty page's activity log shows.
export const useLoyaltyTransactions = (customerId?: number): UseQueryResult<LoyaltyTransaction[]> =>
  useQuery<LoyaltyTransaction[]>({
    queryKey: ['crm-loyalty-transactions', customerId ?? 'all'],
    queryFn: async (): Promise<LoyaltyTransaction[]> => {
      // No customerId = business-wide recent activity feed; the generated
      // client omits the param when undefined.
      const res = await unwrap(crmApi.getLoyaltyTransactions(customerId, 100, 0));
      return z.array(loyaltyTransactionSchema).parse(unwrapList(res, 'transactions'));
    },
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
      await unwrap(crmApi.updateLoyaltyConfig(request));
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
        ? unwrap(crmApi.updateLoyaltyTier(payload.id, request))
        : unwrap(crmApi.createLoyaltyTier(request));
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
        crmApi.upsertLoyaltyReward(
          payload.id,
          payload.data as unknown as Partial<import('../../services/crmApi').LoyaltyReward>
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
      return unwrap(crmApi.awardPoints(payload.customerId, payload.points, payload.reason));
    },
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-members'] });
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
      return unwrap(crmApi.redeemPoints(payload.customerId, payload.points, payload.reason));
    },
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-members'] });
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
