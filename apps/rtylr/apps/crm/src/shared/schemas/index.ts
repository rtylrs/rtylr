export { dashboardStatsSchema, topCustomerSchema, recentCampaignSchema } from './dashboard';
export type { DashboardStats, TopCustomer, RecentCampaign } from './dashboard';

export { customerSchema, timelineEntrySchema } from './customers';
export type { Customer, TimelineEntry } from './customers';

export { segmentSchema } from './segments';
export type { Segment } from './segments';

export { campaignSchema, campaignStatsSchema } from './campaigns';
export type { Campaign, CampaignStats } from './campaigns';

export {
  loyaltyConfigSchema,
  loyaltyTransactionSchema,
  loyaltyMemberSchema,
  loyaltyRewardSchema,
  loyaltyTierSchema,
} from './loyalty';
export type {
  LoyaltyConfig,
  LoyaltyTransaction,
  LoyaltyMember,
  LoyaltyReward,
  LoyaltyTier,
} from './loyalty';

export { automationSchema } from './automations';
export type { Automation } from './automations';

export { templateSchema } from './templates';
export type { Template } from './templates';

export { feedbackItemSchema } from './feedback';
export type { FeedbackItem } from './feedback';

export { settingsSchema } from './settings';
export type { Settings } from './settings';
