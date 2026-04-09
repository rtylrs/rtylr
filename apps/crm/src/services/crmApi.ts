import api from '../lib/api';

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
  getDashboard: () =>
    api.get<DashboardStats>('/crm/dashboard').then((r) => r.data),

  // Customers
  getCustomers: (params?: CustomersParams) =>
    api.get<Customer[]>('/crm/customers', { params }).then((r) => r.data),

  getCustomerDetail: (id: number) =>
    api.get<Customer>('/crm/customers/detail', { params: { id } }).then((r) => r.data),

  getCustomerTimeline: (id: number) =>
    api.get<TimelineEntry[]>('/crm/customers/timeline', { params: { id } }).then((r) => r.data),

  createCustomer: (data: Partial<Customer>) =>
    api.post<Customer>('/crm/customers', data).then((r) => r.data),

  updateCustomer: (id: number, data: Partial<Customer>) =>
    api.post<Customer>('/crm/customers', { id, ...data }).then((r) => r.data),

  deleteCustomer: (id: number) =>
    api.post('/crm/customers', { id, _action: 'delete' }).then((r) => r.data),

  recordVisit: (customerId: number) =>
    api.post('/crm/customers/record-visit', { customerId }).then((r) => r.data),

  importCustomers: (rows: any[]) =>
    api.post('/crm/customers/import', { rows }).then((r) => r.data),

  mergeCustomers: (primaryId: number, secondaryId: number) =>
    api.post('/crm/customers/merge', { primaryId, secondaryId }).then((r) => r.data),

  exportCustomerData: (id: number) =>
    api.get('/crm/customers/export', { params: { id } }).then((r) => r.data),

  // Loyalty
  awardPoints: (customerId: number, points: number, reason?: string) =>
    api.post('/crm/loyalty/award', { customerId, points, reason }).then((r) => r.data),

  redeemPoints: (customerId: number, points: number, reason?: string) =>
    api.post('/crm/loyalty/redeem', { customerId, points, reason }).then((r) => r.data),

  getLoyaltyConfig: () =>
    api.get<LoyaltyConfig>('/crm/loyalty/config').then((r) => r.data),

  updateLoyaltyConfig: (config: Partial<LoyaltyConfig>) =>
    api.post<LoyaltyConfig>('/crm/loyalty/config', config).then((r) => r.data),

  getLoyaltyTransactions: () =>
    api.get<LoyaltyTransaction[]>('/crm/loyalty/transactions').then((r) => r.data),

  getLoyaltyMembers: () =>
    api.get<LoyaltyMember[]>('/crm/loyalty/members').then((r) => r.data),

  getLoyaltyRewards: () =>
    api.get<LoyaltyReward[]>('/crm/loyalty/rewards').then((r) => r.data),

  getLoyaltyTiers: () =>
    api.get<LoyaltyTier[]>('/crm/loyalty/tiers').then((r) => r.data),

  updateLoyaltyTier: (id: number, data: Partial<LoyaltyTier>) =>
    api.post('/crm/loyalty/tiers', { id, ...data }).then((r) => r.data),

  // Segments
  getSegments: () =>
    api.get<Segment[]>('/crm/segments').then((r) => r.data),

  getSegmentById: (id: number) =>
    api.get<Segment>('/crm/segments/detail', { params: { id } }).then((r) => r.data),

  getSegmentCustomers: (id: number) =>
    api.get<Customer[]>('/crm/segments/customers', { params: { id } }).then((r) => r.data),

  createSegment: (data: Partial<Segment>) =>
    api.post<Segment>('/crm/segments', data).then((r) => r.data),

  updateSegment: (id: number, data: Partial<Segment>) =>
    api.post<Segment>('/crm/segments', { id, ...data }).then((r) => r.data),

  deleteSegment: (id: number) =>
    api.post('/crm/segments', { id, _action: 'delete' }).then((r) => r.data),

  // Campaigns
  getCampaigns: () =>
    api.get<Campaign[]>('/crm/campaigns').then((r) => r.data),

  getCampaignById: (id: number) =>
    api.get<Campaign>('/crm/campaigns/detail', { params: { id } }).then((r) => r.data),

  getCampaignStats: (id: number) =>
    api.get<CampaignStats>('/crm/campaigns/stats', { params: { id } }).then((r) => r.data),

  createCampaign: (data: Partial<Campaign>) =>
    api.post<Campaign>('/crm/campaigns', data).then((r) => r.data),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    api.post<Campaign>('/crm/campaigns', { id, ...data }).then((r) => r.data),

  updateCampaignStatus: (id: number, status: string) =>
    api.post('/crm/campaigns/status', { id, status }).then((r) => r.data),

  // Automations
  getAutomations: () =>
    api.get<Automation[]>('/crm/automations').then((r) => r.data),

  createAutomation: (data: Partial<Automation>) =>
    api.post<Automation>('/crm/automations', data).then((r) => r.data),

  updateAutomation: (id: number, data: Partial<Automation>) =>
    api.post<Automation>('/crm/automations', { id, ...data }).then((r) => r.data),

  // Templates
  getTemplates: () =>
    api.get<Template[]>('/crm/templates').then((r) => r.data),

  createTemplate: (data: Partial<Template>) =>
    api.post<Template>('/crm/templates', data).then((r) => r.data),

  // Feedback
  getFeedback: () =>
    api.get<FeedbackItem[]>('/crm/feedback').then((r) => r.data),

  // Settings
  getSettings: () =>
    api.get<Settings>('/crm/settings').then((r) => r.data),

  updateSettings: (data: Partial<Settings>) =>
    api.post<Settings>('/crm/settings', data).then((r) => r.data),
};
