import { z } from 'zod';

// Accept both the backend proto shape (firstName/lastName/totalSpent/
// loyaltyTier) and the legacy display shape, and normalize to what
// Dashboard.tsx renders. The backend has no "visits" field, so it defaults to 0.
export const topCustomerSchema = z
  .object({
    name: z.string().optional(),
    firstName: z.string().optional(),
    lastName: z.string().optional(),
    spend: z.string().optional(),
    totalSpent: z.union([z.number(), z.string()]).optional(),
    visits: z.number().optional(),
    tier: z.string().optional(),
    loyaltyTier: z.string().optional(),
  })
  .transform((c) => ({
    name: c.name ?? `${c.firstName ?? ''} ${c.lastName ?? ''}`.trim(),
    spend: c.spend ?? String(c.totalSpent ?? 0),
    visits: c.visits ?? 0,
    tier: c.tier ?? c.loyaltyTier ?? '',
  }));

export const recentCampaignSchema = z.object({
  name: z.string(),
  channel: z.string(),
  sent: z.number(),
  openRate: z.string(),
  revenue: z.string(),
});

export const dashboardStatsSchema = z.object({
  totalCustomers: z.number(),
  newThisMonth: z.number(),
  activeLoyaltyMembers: z.number(),
  campaignReach: z.number().default(0),
  // Trend strings + recentCampaigns have no backend source yet; default so the
  // dashboard loads instead of throwing on missing fields.
  totalCustomersChange: z.string().default(''),
  newThisMonthChange: z.string().default(''),
  activeLoyaltyMembersChange: z.string().default(''),
  campaignReachChange: z.string().default(''),
  topCustomers: z.array(topCustomerSchema).default([]),
  recentCampaigns: z.array(recentCampaignSchema).default([]),
});

export type DashboardStats = z.infer<typeof dashboardStatsSchema>;
export type TopCustomer = z.infer<typeof topCustomerSchema>;
export type RecentCampaign = z.infer<typeof recentCampaignSchema>;
