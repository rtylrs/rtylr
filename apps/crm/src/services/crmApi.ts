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
}

export interface CustomersParams {
  search?: string;
  tier?: string;
  limit?: number;
  offset?: number;
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

  createCustomer: (data: Partial<Customer>) =>
    api.post<Customer>('/crm/customers', data).then((r) => r.data),

  updateCustomer: (id: number, data: Partial<Customer>) =>
    api.post<Customer>('/crm/customers', { id, ...data }).then((r) => r.data),

  deleteCustomer: (id: number) =>
    api.post('/crm/customers', { id, _action: 'delete' }).then((r) => r.data),

  recordVisit: (customerId: number) =>
    api.post('/crm/customers/record-visit', { customerId }).then((r) => r.data),

  // Loyalty
  awardPoints: (customerId: number, points: number) =>
    api.post('/crm/loyalty/award', { customerId, points }).then((r) => r.data),

  redeemPoints: (customerId: number, points: number) =>
    api.post('/crm/loyalty/redeem', { customerId, points }).then((r) => r.data),

  getLoyaltyConfig: () =>
    api.get<LoyaltyConfig>('/crm/loyalty/config').then((r) => r.data),

  updateLoyaltyConfig: (config: Partial<LoyaltyConfig>) =>
    api.post<LoyaltyConfig>('/crm/loyalty/config', config).then((r) => r.data),

  getLoyaltyTransactions: () =>
    api.get<LoyaltyTransaction[]>('/crm/loyalty/transactions').then((r) => r.data),

  // Segments
  getSegments: () =>
    api.get<Segment[]>('/crm/segments').then((r) => r.data),

  createSegment: (data: Partial<Segment>) =>
    api.post<Segment>('/crm/segments', data).then((r) => r.data),

  updateSegment: (id: number, data: Partial<Segment>) =>
    api.post<Segment>('/crm/segments', { id, ...data }).then((r) => r.data),

  deleteSegment: (id: number) =>
    api.post('/crm/segments', { id, _action: 'delete' }).then((r) => r.data),

  // Campaigns
  getCampaigns: () =>
    api.get<Campaign[]>('/crm/campaigns').then((r) => r.data),

  createCampaign: (data: Partial<Campaign>) =>
    api.post<Campaign>('/crm/campaigns', data).then((r) => r.data),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    api.post<Campaign>('/crm/campaigns', { id, ...data }).then((r) => r.data),

  updateCampaignStatus: (id: number, status: string) =>
    api.post('/crm/campaigns/status', { id, status }).then((r) => r.data),
};
