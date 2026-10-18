import { z } from 'zod';

/**
 * OneKit int64 fields travel as strings on the wire (docs/cross-app/07-plumbing.md).
 * Pass the value through unchanged on failure so validation still fails loudly
 * instead of masking bad input as 0.
 */
const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

const legacyLoyaltyConfigSchema = z.object({
  programEnabled: z.boolean(),
  pointsPerDollar: z.number(),
  pointsExpiry: z.number(),
  redemptionRate: z.number(),
});

// The wire omits a field holding its zero value (false, 0), so an absent field
// is that zero value, not a missing one: a paused program arrives with no
// is_enabled at all. Requiring the field failed every paused config, and
// substituting a default showed a saved 0% discount as 5%. The server sends
// real defaults for a business that never saved settings.
const backendLoyaltyConfigSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number).optional(),
  isEnabled: z.boolean().default(false),
  pointsPerCurrencyUnit: finiteNumber.default(0),
  currencyPerPoint: finiteNumber.default(0),
  bronzeMin: finiteNumber.default(0),
  silverMin: finiteNumber.default(0),
  goldMin: finiteNumber.default(0),
  platinumMin: finiteNumber.default(0),
  bronzeDiscountPct: finiteNumber.default(0),
  silverDiscountPct: finiteNumber.default(0),
  goldDiscountPct: finiteNumber.default(0),
  platinumDiscountPct: finiteNumber.default(0),
  pointsExpiryDays: finiteNumber.default(0),
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
  id: z.union([z.string(), z.number()]).transform(Number).optional(),
  customer: z.string(),
  action: z.string(),
  time: z.string(),
  type: z.enum(['earn', 'redeem', 'upgrade', 'adjust', 'expire', 'award']).catch('earn'),
});

const backendLoyaltyTransactionSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  customerId: z.union([z.string(), z.number()]).transform(Number),
  type: z.enum(['earn', 'redeem', 'adjust', 'expire', 'award']),
  points: finiteNumber.catch(0),
  balanceBefore: finiteNumber.catch(0),
  balanceAfter: finiteNumber.catch(0),
  description: z.string().nullable().optional(),
  createdAt: z.string(),
});

export const loyaltyTransactionSchema = z
  .union([legacyLoyaltyTransactionSchema, backendLoyaltyTransactionSchema])
  .transform((value) => {
    if ('customerId' in value) {
      return {
        id: value.id,
        customer: String(value.customerId),
        action: value.description || value.type,
        time: value.createdAt,
        type: value.type,
        customerId: value.customerId,
        points: value.points,
        balanceBefore: value.balanceBefore,
        balanceAfter: value.balanceAfter,
      };
    }
    return { ...value, id: value.id ?? 0 };
  });

export const loyaltyMemberSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  initials: z.string(),
  points: finiteNumber,
  tier: z.string(),
  visits: finiteNumber,
  joinedAt: z.string(),
  lastActive: z.string(),
});

export const loyaltyRewardSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string().catch(''),
  pointsCost: finiteNumber.catch(0),
  category: z.string().catch('general'),
  stock: z
    .union([z.number(), z.string(), z.null(), z.undefined()])
    .transform((val) => {
      if (val == null) return null;
      const num = Number(val);
      return Number.isNaN(num) ? null : num;
    })
    .nullable()
    .catch(null),
  active: z.boolean().catch(true),
  icon: z.string().nullable().optional(),
});

const legacyLoyaltyTierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number).optional(),
  tierId: z.union([z.string(), z.number()]).transform(Number).optional(),
  tier_id: z.union([z.string(), z.number()]).transform(Number).optional(),
  name: z.string(),
  threshold: finiteNumber.optional(),
  minPoints: finiteNumber.optional(),
  min_points: finiteNumber.optional(),
  multiplier: z.number().optional(),
  discountPct: z.number().optional(),
  discount_pct: z.number().optional(),
  perks: z.union([z.string(), z.array(z.string())]).optional(),
  color: z.string().optional(),
  sortOrder: finiteNumber.optional(),
  sort_order: finiteNumber.optional(),
});

const backendLoyaltyTierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number).optional(),
  tierId: z.union([z.string(), z.number()]).transform(Number).optional(),
  tier_id: z.union([z.string(), z.number()]).transform(Number).optional(),
  name: z.string(),
  minPoints: finiteNumber.optional(),
  min_points: finiteNumber.optional(),
  discountPct: z.number().optional(),
  discount_pct: z.number().optional(),
  perks: z.union([z.string(), z.array(z.string())]).optional(),
  sortOrder: finiteNumber.optional(),
  sort_order: finiteNumber.optional(),
});

export interface LoyaltyTier {
  id: number;
  name: string;
  threshold: number;
  multiplier: number;
  discountPct: number;
  perks: string[];
  color: string;
  sortOrder: number;
}

export const loyaltyTierSchema = z
  .union([legacyLoyaltyTierSchema, backendLoyaltyTierSchema])
  .transform((value): LoyaltyTier => {
    const minPoints =
      'threshold' in value && value.threshold !== undefined
        ? value.threshold
        : (value.minPoints ?? value.min_points ?? 0);
    const discountPct =
      'multiplier' in value && value.multiplier !== undefined
        ? value.multiplier
        : (value.discountPct ?? value.discount_pct ?? 0);
    const sortOrder =
      'sortOrder' in value && value.sortOrder !== undefined
        ? value.sortOrder
        : (value.sort_order ?? 0);
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
    const id =
      'id' in value && value.id !== undefined
        ? value.id
        : 'tierId' in value && value.tierId !== undefined
          ? value.tierId
          : 'tier_id' in value && value.tier_id !== undefined
            ? value.tier_id
            : 0;
    return {
      id,
      name: value.name,
      threshold: minPoints,
      multiplier: discountPct,
      discountPct,
      perks,
      color: 'color' in value && value.color ? value.color : (colors[value.name] ?? '#3b82f6'),
      sortOrder,
    };
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

// -- Domain constants ---------------------------------------------------------

export const LOYALTY_TRANSACTION_TYPES = ['earn', 'redeem', 'upgrade'] as const;
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

export function getTransactionLabel(type: LoyaltyTransaction['type']): string {
  if (type === 'earn') return 'Points Earned';
  if (type === 'redeem') return 'Points Redeemed';
  if (type === 'award') return 'Points Awarded';
  if (type === 'adjust') return 'Points Adjusted';
  if (type === 'expire') return 'Points Expired';
  if (type === 'upgrade') return 'Tier Upgrade';
  return String(type);
}
