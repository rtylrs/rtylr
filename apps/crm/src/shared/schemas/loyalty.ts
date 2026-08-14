import { z } from 'zod';

const legacyLoyaltyConfigSchema = z.object({
  programEnabled: z.boolean(),
  pointsPerDollar: z.number(),
  pointsExpiry: z.number(),
  redemptionRate: z.number(),
});

const backendLoyaltyConfigSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number).optional(),
  isEnabled: z.boolean(),
  pointsPerCurrencyUnit: z.number(),
  currencyPerPoint: z.number(),
  bronzeMin: z.number().catch(0),
  silverMin: z.number().catch(500),
  goldMin: z.number().catch(1500),
  platinumMin: z.number().catch(5000),
  bronzeDiscountPct: z.number().catch(0),
  silverDiscountPct: z.number().catch(5),
  goldDiscountPct: z.number().catch(10),
  platinumDiscountPct: z.number().catch(15),
  pointsExpiryDays: z.number().catch(365),
  updatedAt: z.string().optional(),
});

export const loyaltyConfigSchema = z
  .union([legacyLoyaltyConfigSchema, backendLoyaltyConfigSchema])
  .transform((value) => {
    if ('isEnabled' in value) {
      return {
        programEnabled: value.isEnabled,
        pointsPerDollar: value.pointsPerCurrencyUnit,
        pointsExpiry: value.pointsExpiryDays,
        redemptionRate: value.currencyPerPoint,
        bronzeMin: value.bronzeMin,
        silverMin: value.silverMin,
        goldMin: value.goldMin,
        platinumMin: value.platinumMin,
        bronzeDiscountPct: value.bronzeDiscountPct,
        silverDiscountPct: value.silverDiscountPct,
        goldDiscountPct: value.goldDiscountPct,
        platinumDiscountPct: value.platinumDiscountPct,
        updatedAt: value.updatedAt,
      };
    }
    return value;
  });

const legacyLoyaltyTransactionSchema = z.object({
  customer: z.string(),
  action: z.string(),
  time: z.string(),
  type: z.enum(['earn', 'redeem', 'upgrade', 'adjust', 'expire', 'award']).catch('earn'),
});

const backendLoyaltyTransactionSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  customerId: z.union([z.string(), z.number()]).transform(Number),
  type: z.enum(['earn', 'redeem', 'adjust', 'expire', 'award']),
  points: z.number().catch(0),
  balanceBefore: z.number().catch(0),
  balanceAfter: z.number().catch(0),
  description: z.string().nullable().optional(),
  createdAt: z.string(),
});

export const loyaltyTransactionSchema = z
  .union([legacyLoyaltyTransactionSchema, backendLoyaltyTransactionSchema])
  .transform((value) => {
    if ('customerId' in value) {
      return {
        customer: String(value.customerId),
        action: value.description || value.type,
        time: value.createdAt,
        type: value.type,
        id: value.id,
        customerId: value.customerId,
        points: value.points,
        balanceBefore: value.balanceBefore,
        balanceAfter: value.balanceAfter,
      };
    }
    return value;
  });

export const loyaltyMemberSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  initials: z.string(),
  points: z.number(),
  tier: z.string(),
  visits: z.number(),
  joinedAt: z.string(),
  lastActive: z.string(),
});

export const loyaltyRewardSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string(),
  pointsCost: z.number(),
  category: z.string(),
  stock: z.number().nullable(),
  active: z.boolean(),
  icon: z.string().nullable().optional(),
});

const legacyLoyaltyTierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  threshold: z.number(),
  multiplier: z.number(),
  perks: z.array(z.string()).default([]),
  color: z.string(),
});

const backendLoyaltyTierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  minPoints: z.number().catch(0),
  discountPct: z.number().catch(0),
  perks: z.union([z.string(), z.array(z.string())]).optional(),
  sortOrder: z.number().catch(0),
});

export const loyaltyTierSchema = z
  .union([legacyLoyaltyTierSchema, backendLoyaltyTierSchema])
  .transform((value) => {
    if ('minPoints' in value) {
      let perks: string[] = [];
      if (Array.isArray(value.perks)) perks = value.perks;
      if (typeof value.perks === 'string') {
        try {
          const parsed = JSON.parse(value.perks) as unknown;
          if (Array.isArray(parsed))
            perks = parsed.filter((item): item is string => typeof item === 'string');
        } catch {
          perks = value.perks ? [value.perks] : [];
        }
      }
      const colors: Record<string, string> = {
        Bronze: '#a16207',
        Silver: '#64748b',
        Gold: '#f59e0b',
        Platinum: '#8b5cf6',
      };
      return {
        id: value.id,
        name: value.name,
        threshold: value.minPoints,
        multiplier: value.discountPct,
        discountPct: value.discountPct,
        perks,
        color: colors[value.name] ?? '#3b82f6',
        sortOrder: value.sortOrder,
      };
    }
    return value;
  });

export interface LoyaltyConfig {
  programEnabled: boolean;
  pointsPerDollar: number;
  pointsExpiry: number;
  redemptionRate: number;
  bronzeMin?: number;
  silverMin?: number;
  goldMin?: number;
  platinumMin?: number;
  bronzeDiscountPct?: number;
  silverDiscountPct?: number;
  goldDiscountPct?: number;
  platinumDiscountPct?: number;
  updatedAt?: string;
}
export type LoyaltyTransaction = z.infer<typeof loyaltyTransactionSchema>;
export type LoyaltyMember = z.infer<typeof loyaltyMemberSchema>;
export type LoyaltyReward = z.infer<typeof loyaltyRewardSchema>;
export interface LoyaltyTier {
  id: number;
  name: string;
  threshold: number;
  multiplier: number;
  discountPct?: number;
  perks: string[];
  color: string;
  sortOrder?: number;
}

// -- Domain constants ---------------------------------------------------------

export const LOYALTY_TRANSACTION_TYPES = ['earn', 'redeem', 'upgrade'] as const;
export type LoyaltyTransactionType =
  | (typeof LOYALTY_TRANSACTION_TYPES)[number]
  | 'adjust'
  | 'expire'
  | 'award';

// -- Helpers ------------------------------------------------------------------

export function isEarnTransaction(tx: LoyaltyTransaction): boolean {
  return tx.type === 'earn';
}

export function isRedeemTransaction(tx: LoyaltyTransaction): boolean {
  return tx.type === 'redeem';
}

export function getTransactionIcon(type: LoyaltyTransaction['type']): string {
  if (type === 'earn') return 'plus';
  if (type === 'redeem') return 'minus';
  return 'arrow-up';
}

export function isUpgradeTransaction(tx: LoyaltyTransaction): boolean {
  return tx.type === 'upgrade';
}

export function getTransactionLabel(type: LoyaltyTransaction['type']): string {
  if (type === 'earn') return 'Points Earned';
  if (type === 'redeem') return 'Points Redeemed';
  return 'Tier Upgrade';
}

export function isActiveMember(member: LoyaltyMember, atRiskDays = 90): boolean {
  const last = new Date(member.lastActive);
  const now = new Date();
  const diffMs = now.getTime() - last.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  return diffDays <= atRiskDays;
}

export function memberInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}
