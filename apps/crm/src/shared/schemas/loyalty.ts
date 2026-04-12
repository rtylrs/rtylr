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
