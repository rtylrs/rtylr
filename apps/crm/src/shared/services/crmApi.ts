import { safe } from '../lib/api';
import { unwrap } from '@voxire/api';

// ── Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  totalCustomers: number;
  newThisMonth: number;
  activeLoyaltyMembers: number;
  campaignReach: number;
  totalCustomersChange: string;
  newThisMonthChange: string;
  activeLoyaltyMembersChange: string;
  campaignReachChange: string;
  topCustomers: TopCustomer[];
  recentCampaigns: RecentCampaign[];
}

export interface TopCustomer {
  name: string;
  spend: string;
  visits: number;
  tier: string;
}

export interface RecentCampaign {
  name: string;
  channel: string;
  sent: number;
  openRate: string;
  revenue: string;
}

export interface Customer {
  id: number;
  name: string;
  email: string;
  phone: string;
  totalSpent: number;
  visits: number;
  loyaltyPoints: number;
  lastVisit: string;
  initials: string;
  tags?: string[];
  notes?: string;
  birthday?: string;
  consent?: { email: boolean; sms: boolean; marketing: boolean };
  createdAt?: string;
  referredBy?: number | null;
}

export interface CustomersParams {
  search?: string;
  tier?: string;
  limit?: number;
  offset?: number;
}

export interface TimelineEntry {
  id: number;
  type: 'visit' | 'purchase' | 'email' | 'sms' | 'note' | 'points' | 'reward';
  title: string;
  description?: string;
  at: string;
  meta?: Record<string, any>;
}

export interface Segment {
  id: number;
  name: string;
  description: string;
  customerCount: number;
  color: string;
  rules: string[];
  trend: string;
  trendUp: boolean;
}

export interface SegmentRuleObj {
  field: string;
  operator: string;
  value: string | number;
}

export interface Campaign {
  id: number;
  name: string;
  description: string;
  channel: string;
  status: string;
  audience: number;
  sent: number;
  openRate: string;
  clickRate: string;
  revenue: string;
  updatedAt: string;
  subject?: string;
  body?: string;
  segmentId?: number;
  scheduledFor?: string;
}

export interface CampaignStats {
  id: number;
  sent: number;
  delivered: number;
  opened: number;
  clicked: number;
  unsubscribed: number;
  bounced: number;
  revenue: number;
  timeline: Array<{ day: string; opens: number; clicks: number }>;
}

export interface LoyaltyConfig {
  programEnabled: boolean;
  pointsPerDollar: number;
  pointsExpiry: number;
  redemptionRate: number;
}

export interface LoyaltyTransaction {
  customer: string;
  action: string;
  time: string;
  type: 'earn' | 'redeem' | 'upgrade';
}

export interface LoyaltyMember {
  id: number;
  name: string;
  initials: string;
  points: number;
  tier: string;
  visits: number;
  joinedAt: string;
  lastActive: string;
}

export interface LoyaltyReward {
  id: number;
  name: string;
  description: string;
  pointsCost: number;
  category: string;
  stock: number;
  active: boolean;
  icon?: string;
}

export interface LoyaltyTier {
  id: number;
  name: string;
  threshold: number;
  multiplier: number;
  perks: string[];
  color: string;
}

export interface Automation {
  id: number;
  name: string;
  description: string;
  trigger: string;
  status: 'active' | 'paused' | 'draft';
  enrolled: number;
  completed: number;
  createdAt: string;
  nodes?: Array<{ id: string; type: string; label: string; x: number; y: number }>;
}

export interface Template {
  id: number;
  name: string;
  channel: 'email' | 'sms';
  category: string;
  subject?: string;
  body: string;
  thumbnail?: string;
  updatedAt: string;
}

export interface FeedbackItem {
  id: number;
  customer: string;
  rating: number;
  nps: number;
  comment: string;
  createdAt: string;
  source: string;
}

export interface Settings {
  defaultConsent: boolean;
  timezone: string;
  fromEmail: string;
  fromName: string;
  smsSenderId: string;
  businessName: string;
  atRiskDays: number;
  currency: string;
}

// ── Service ──────────────────────────────────────────────────────

export const crmApi = {
  // Dashboard
  getDashboard: () => safe.get<DashboardStats>('/crm/dashboard'),

  // Customers
  getCustomers: (params?: CustomersParams) => safe.get<Customer[]>('/crm/customers', { params }),

  getCustomerDetail: (id: number) =>
    safe.get<Customer>('/crm/customers/detail', { params: { id } }),

  getCustomerTimeline: (id: number) =>
    safe.get<TimelineEntry[]>('/crm/customers/timeline', { params: { id } }),

  createCustomer: (data: Partial<Customer>) => safe.post<Customer>('/crm/customers', data),

  updateCustomer: (id: number, data: Partial<Customer>) =>
    safe.post<Customer>('/crm/customers', { id, ...data }),

  deleteCustomer: (id: number) => safe.post('/crm/customers', { id, _action: 'delete' }),

  recordVisit: (customerId: number) => safe.post('/crm/customers/record-visit', { customerId }),

  importCustomers: (rows: Record<string, unknown>[]) =>
    safe.post('/crm/customers/import', { rows }),

  mergeCustomers: (primaryId: number, secondaryId: number) =>
    safe.post('/crm/customers/merge', { primaryId, secondaryId }),

  exportCustomerData: (id: number) => safe.get('/crm/customers/export', { params: { id } }),

  // Loyalty
  awardPoints: (customerId: number, points: number, reason?: string) =>
    safe.post('/crm/loyalty/award', { customerId, points, reason }),

  redeemPoints: (customerId: number, points: number, reason?: string) =>
    safe.post('/crm/loyalty/redeem', { customerId, points, reason }),

  getLoyaltyConfig: () => safe.get<LoyaltyConfig>('/crm/loyalty/config'),

  updateLoyaltyConfig: (config: Partial<LoyaltyConfig>) =>
    safe.post<LoyaltyConfig>('/crm/loyalty/config', config),

  getLoyaltyTransactions: () => safe.get<LoyaltyTransaction[]>('/crm/loyalty/transactions'),

  getLoyaltyMembers: () => safe.get<LoyaltyMember[]>('/crm/loyalty/members'),

  getLoyaltyRewards: () => safe.get<LoyaltyReward[]>('/crm/loyalty/rewards'),

  getLoyaltyTiers: () => safe.get<LoyaltyTier[]>('/crm/loyalty/tiers'),

  updateLoyaltyTier: (id: number, data: Partial<LoyaltyTier>) =>
    safe.post('/crm/loyalty/tiers', { id, ...data }),

  // Segments
  getSegments: () => safe.get<Segment[]>('/crm/segments'),

  getSegmentById: (id: number) => safe.get<Segment>('/crm/segments/detail', { params: { id } }),

  getSegmentCustomers: (id: number) =>
    safe.get<Customer[]>('/crm/segments/customers', { params: { id } }),

  createSegment: (data: Partial<Segment>) => safe.post<Segment>('/crm/segments', data),

  updateSegment: (id: number, data: Partial<Segment>) =>
    safe.post<Segment>('/crm/segments', { id, ...data }),

  deleteSegment: (id: number) => safe.post('/crm/segments', { id, _action: 'delete' }),

  // Campaigns
  getCampaigns: () => safe.get<Campaign[]>('/crm/campaigns'),

  getCampaignById: (id: number) => safe.get<Campaign>('/crm/campaigns/detail', { params: { id } }),

  getCampaignStats: (id: number) =>
    safe.get<CampaignStats>('/crm/campaigns/stats', { params: { id } }),

  createCampaign: (data: Partial<Campaign>) => safe.post<Campaign>('/crm/campaigns', data),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    safe.post<Campaign>('/crm/campaigns', { id, ...data }),

  updateCampaignStatus: (id: number, status: string) =>
    safe.post('/crm/campaigns/status', { id, status }),

  deleteCampaign: (id: number) => safe.post('/crm/campaigns', { id, _action: 'delete' }),

  // Automations
  getAutomations: () => safe.get<Automation[]>('/crm/automations'),

  createAutomation: (data: Partial<Automation>) => safe.post<Automation>('/crm/automations', data),

  updateAutomation: (id: number, data: Partial<Automation>) =>
    safe.post<Automation>('/crm/automations', { id, ...data }),

  // Templates
  getTemplates: () => safe.get<Template[]>('/crm/templates'),

  createTemplate: (data: Partial<Template>) => safe.post<Template>('/crm/templates', data),

  // Feedback
  getFeedback: () => safe.get<FeedbackItem[]>('/crm/feedback'),

  // Settings
  getSettings: () => safe.get<Settings>('/crm/settings'),

  updateSettings: (data: Partial<Settings>) => safe.post<Settings>('/crm/settings', data),
};
