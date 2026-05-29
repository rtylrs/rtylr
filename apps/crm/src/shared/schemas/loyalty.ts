import { z } from 'zod';

export const loyaltyConfigSchema = z.object({
  programEnabled: z.boolean(),
  pointsPerDollar: z.number(),
  pointsExpiry: z.number(),
  redemptionRate: z.number(),
});

export const loyaltyTransactionSchema = z.object({
  customer: z.string(),
  action: z.string(),
  time: z.string(),
  type: z.enum(['earn', 'redeem', 'upgrade']).catch('earn'),
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
  stock: z.number(),
  active: z.boolean(),
  icon: z.string().nullable().optional(),
});

export const loyaltyTierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  threshold: z.number(),
  multiplier: z.number(),
  perks: z.array(z.string()).default([]),
  color: z.string(),
});

export type LoyaltyConfig = z.infer<typeof loyaltyConfigSchema>;
export type LoyaltyTransaction = z.infer<typeof loyaltyTransactionSchema>;
export type LoyaltyMember = z.infer<typeof loyaltyMemberSchema>;
export type LoyaltyReward = z.infer<typeof loyaltyRewardSchema>;
export type LoyaltyTier = z.infer<typeof loyaltyTierSchema>;

// -- Domain constants ---------------------------------------------------------

export const LOYALTY_TRANSACTION_TYPES = ['earn', 'redeem', 'upgrade'] as const;
export type LoyaltyTransactionType = (typeof LOYALTY_TRANSACTION_TYPES)[number];

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
