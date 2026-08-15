import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

const optionalFiniteNumber = z.preprocess((value) => {
  if (value === undefined || value === null || value === '') return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}, z.number().optional());

// Accept both the backend proto shape (firstName/lastName/totalSpent/
// loyaltyTier) and the legacy display shape, and normalize to what
// Dashboard.tsx renders. The backend has no "visits" field, so it remains
// undefined instead of being presented as a real zero.
export const topCustomerSchema = z
  .object({
    name: z.string().optional(),
    firstName: z.string().optional(),
    lastName: z.string().optional(),
    spend: z.string().optional(),
    totalSpent: z.union([z.number(), z.string()]).optional(),
    visits: finiteNumber.optional(),
    tier: z.string().optional(),
    loyaltyTier: z.string().optional(),
  })
  .transform((c) => ({
    name: c.name ?? `${c.firstName ?? ''} ${c.lastName ?? ''}`.trim(),
    spend: c.spend ?? String(c.totalSpent ?? 0),
    visits: c.visits,
    tier: c.tier ?? c.loyaltyTier ?? '',
  }));

export const recentCampaignSchema = z.object({
  name: z.string(),
  channel: z.string(),
  sent: finiteNumber,
  openRate: z.string(),
  revenue: z.string(),
});

export const dashboardStatsSchema = z
  .object({
    totalCustomers: finiteNumber.optional(),
    total_customers: finiteNumber.optional(),
    newThisMonth: finiteNumber.optional(),
    new_this_month: finiteNumber.optional(),
    activeLoyaltyMembers: finiteNumber.optional(),
    active_loyalty_members: finiteNumber.optional(),
    campaignReach: optionalFiniteNumber,
    campaign_reach: optionalFiniteNumber,
    // Trend strings + recentCampaigns have no backend source yet; default so the
    // dashboard loads instead of throwing on missing fields.
    totalCustomersChange: z.string().optional(),
    total_customers_change: z.string().optional(),
    newThisMonthChange: z.string().optional(),
    new_this_month_change: z.string().optional(),
    activeLoyaltyMembersChange: z.string().optional(),
    active_loyalty_members_change: z.string().optional(),
    campaignReachChange: z.string().optional(),
    campaign_reach_change: z.string().optional(),
    topCustomers: z.array(topCustomerSchema).optional(),
    top_customers: z.array(topCustomerSchema).optional(),
    recentCampaigns: z.array(recentCampaignSchema).optional(),
    recent_campaigns: z.array(recentCampaignSchema).optional(),
  })
  .transform((raw) => ({
    totalCustomers: raw.totalCustomers ?? raw.total_customers ?? 0,
    newThisMonth: raw.newThisMonth ?? raw.new_this_month ?? 0,
    activeLoyaltyMembers: raw.activeLoyaltyMembers ?? raw.active_loyalty_members ?? 0,
    // The CRM dashboard endpoint does not currently provide campaign reach.
    // Keep it undefined so the UI can say "Unavailable" instead of presenting
    // a fabricated zero as a real metric.
    campaignReach: raw.campaignReach ?? raw.campaign_reach,
    totalCustomersChange: raw.totalCustomersChange ?? raw.total_customers_change ?? '',
    newThisMonthChange: raw.newThisMonthChange ?? raw.new_this_month_change ?? '',
    activeLoyaltyMembersChange:
      raw.activeLoyaltyMembersChange ?? raw.active_loyalty_members_change ?? '',
    campaignReachChange: raw.campaignReachChange ?? raw.campaign_reach_change ?? '',
    topCustomers: raw.topCustomers ?? raw.top_customers ?? [],
    recentCampaigns: raw.recentCampaigns ?? raw.recent_campaigns ?? [],
  }));

export type DashboardStats = z.infer<typeof dashboardStatsSchema>;
export type TopCustomer = z.infer<typeof topCustomerSchema>;
export type RecentCampaign = z.infer<typeof recentCampaignSchema>;
