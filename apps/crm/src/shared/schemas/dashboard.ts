import { z } from 'zod';

export const topCustomerSchema = z.object({
  name: z.string(),
  spend: z.string(),
  visits: z.number(),
  tier: z.string(),
});

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
  campaignReach: z.number(),
  totalCustomersChange: z.string(),
  newThisMonthChange: z.string(),
  activeLoyaltyMembersChange: z.string(),
  campaignReachChange: z.string(),
  topCustomers: z.array(topCustomerSchema).default([]),
  recentCampaigns: z.array(recentCampaignSchema).default([]),
});

export type DashboardStats = z.infer<typeof dashboardStatsSchema>;
export type TopCustomer = z.infer<typeof topCustomerSchema>;
export type RecentCampaign = z.infer<typeof recentCampaignSchema>;
