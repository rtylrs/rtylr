import { safe } from '../lib/api';
import { toOptionalFiniteNumber, unwrap } from '@voxire/api';

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
  /** Canonical persisted tier from the CRM customer API. */
  loyaltyTier?: string;
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
  type:
    | 'visit'
    | 'purchase'
    | 'email'
    | 'sms'
    | 'note'
    | 'points'
    | 'reward'
    | 'adjust'
    | 'expire'
    | 'award'
    | 'upgrade';
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

// ── Sales Pipeline Types ─────────────────────────────────────────

export type LeadStage =
  | 'cold'
  | 'contacted'
  | 'demo_booked'
  | 'demo_done'
  | 'proposal_sent'
  | 'contracted'
  | 'lost';

export type TouchpointType = 'phone' | 'email' | 'meeting';
export type TouchpointOutcome = 'positive' | 'neutral' | 'negative' | 'no_answer';
export type ContractPlan = 'free' | 'pro' | 'scale';

export interface Lead {
  id: number;
  // API returns company_name → companyName; older records may use company
  company?: string;
  companyName?: string;
  industry?: string;
  size?: string;
  source?: string;
  stage: LeadStage;
  // API returns deal_value → dealValue; mrrPotential kept for backwards compat
  mrrPotential?: number;
  dealValue?: number;
  // API returns assigned_to → assignedTo (user ID, may be null)
  assignedTo?: number | null;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PipelineColumn {
  stage: LeadStage;
  // The backend contract exposes the number of leads as `count`.
  count: number;
  totalMrr: number;
}

export interface Touchpoint {
  id: number;
  leadId: number;
  type: TouchpointType;
  outcome: TouchpointOutcome;
  notes: string;
  repName: string;
  contactId: number | null;
  nextFollowUp: string | null;
  createdAt: string;
}

export interface SalesContact {
  id: number;
  leadId: number;
  name: string;
  role: string;
  phone: string;
  email: string;
  createdAt: string;
}

export interface Contract {
  id: number;
  leadId: number;
  company: string;
  plan: ContractPlan;
  mrr: number;
  seats: number;
  signedAt: string;
  signedBy: string;
  closedBy: string;
  notes: string;
  createdAt: string;
}

export interface CreateLeadRequest {
  company: string;
  companyName?: string;
  industry: string;
  size: string;
  source: string;
  stage: LeadStage;
  mrrPotential: number;
  assignedTo?: number;
  notes?: string;
}

export interface UpdateLeadRequest {
  company?: string;
  companyName?: string;
  industry?: string;
  size?: string;
  source?: string;
  stage?: LeadStage;
  mrrPotential?: number;
  assignedTo?: number;
  notes?: string;
}

interface LeadMutationPayload {
  companyName: string;
  industry?: string;
  size?: string;
  source?: string;
  stage?: LeadStage;
  assignedTo?: number;
  mrrPotential?: number;
  notes?: string;
}

export const toLeadMutationPayload = (
  data: CreateLeadRequest | UpdateLeadRequest
): LeadMutationPayload => ({
  companyName: data.companyName ?? data.company ?? '',
  industry: data.industry,
  size: data.size,
  source: data.source,
  stage: data.stage,
  assignedTo: data.assignedTo,
  mrrPotential: data.mrrPotential,
  notes: data.notes,
});

interface CustomerMutationPayload {
  firstName: string;
  lastName: string;
  email?: string;
  phone?: string;
  dateOfBirth?: string;
  tags?: string[];
  notes?: string;
}

function splitCustomerName(name?: string): { firstName: string; lastName: string } {
  const parts = (name ?? '').trim().split(/\s+/).filter(Boolean);
  return {
    firstName: parts[0] ?? '',
    lastName: parts.slice(1).join(' '),
  };
}

function toCustomerMutationPayload(data: Partial<Customer>): CustomerMutationPayload {
  const names = splitCustomerName(data.name);
  return {
    firstName: names.firstName,
    lastName: names.lastName,
    email: data.email,
    phone: data.phone,
    dateOfBirth: data.birthday,
    tags: data.tags,
    notes: data.notes,
  };
}

interface ImportCustomersPayload {
  customers: Record<string, unknown>[];
}

interface LoyaltyPointsPayload {
  customerId: number;
  points: number;
  description?: string;
}

interface LoyaltyConfigPayload {
  pointsPerDollar?: number;
  pointsValue?: number;
  isActive?: boolean;
}

function toLoyaltyConfigPayload(config: Partial<LoyaltyConfig>): LoyaltyConfigPayload {
  return {
    pointsPerDollar: config.pointsPerDollar,
    pointsValue: config.redemptionRate,
    isActive: config.programEnabled,
  };
}

interface SegmentMutationPayload {
  name?: string;
  description?: string;
  filters?: { rules: string[] };
}

function toSegmentMutationPayload(data: Partial<Segment>): SegmentMutationPayload {
  return {
    name: data.name,
    description: data.description,
    filters: { rules: data.rules ?? [] },
  };
}

interface CampaignMutationPayload {
  name?: string;
  type?: string;
  audienceSegmentId?: number;
  subject?: string;
  body?: string;
  scheduledAt?: string;
}

function toCampaignMutationPayload(data: Partial<Campaign>): CampaignMutationPayload {
  return {
    name: data.name,
    type: data.channel,
    audienceSegmentId: data.segmentId,
    subject: data.subject,
    body: data.body,
    scheduledAt: data.scheduledFor,
  };
}

interface AutomationMutationPayload {
  name?: string;
  triggerType?: string;
  triggerConfig?: Record<string, unknown>;
  actions?: Array<{ id: string; type: string; label: string; x: number; y: number }>;
  isActive?: boolean;
}

function toAutomationMutationPayload(data: Partial<Automation>): AutomationMutationPayload {
  return {
    name: data.name,
    triggerType: data.trigger,
    triggerConfig: {},
    actions: data.nodes,
    isActive: data.status === 'active',
  };
}

interface SettingsMutationPayload {
  defaultConsent?: boolean;
  timezone?: string;
  fromEmail?: string;
  fromName?: string;
  smsSenderId?: string;
}

function toSettingsMutationPayload(data: Partial<Settings>): SettingsMutationPayload {
  return {
    defaultConsent: data.defaultConsent,
    timezone: data.timezone,
    fromEmail: data.fromEmail,
    fromName: data.fromName,
    smsSenderId: data.smsSenderId,
  };
}

export interface CreateTouchpointRequest {
  leadId: number;
  type: TouchpointType;
  outcome: TouchpointOutcome;
  notes: string;
  repName: string;
  contactId?: number | null;
  nextFollowUp?: string | null;
}

export interface CreateContactRequest {
  leadId: number;
  name: string;
  role: string;
  phone: string;
  email: string;
}

export interface UpdateContactRequest {
  name?: string;
  role?: string;
  phone?: string;
  email?: string;
}

export interface CreateContractRequest {
  leadId: number;
  plan: ContractPlan;
  mrr: number;
  seats: number;
  signedAt: string;
  signedBy: string;
  closedBy: string;
  notes?: string;
}

export interface UpdateContractRequest {
  plan?: ContractPlan;
  mrr?: number;
  seats?: number;
  signedAt?: string;
  signedBy?: string;
  closedBy?: string;
  notes?: string;
}

interface TouchpointMutationPayload {
  leadId: number;
  type: TouchpointType;
  outcome: TouchpointOutcome;
  notes: string;
  contactId?: number | null;
  nextFollowUp?: string | null;
}

function toTouchpointMutationPayload(data: CreateTouchpointRequest): TouchpointMutationPayload {
  return {
    leadId: data.leadId,
    type: data.type,
    outcome: data.outcome,
    notes: data.notes,
    contactId: data.contactId,
    nextFollowUp: data.nextFollowUp,
  };
}

interface ContractMutationPayload {
  plan?: ContractPlan;
  mrr?: number;
  seats?: number;
  signedAt?: string;
  signedByName?: string;
  closedBy?: number;
  notes?: string;
}

interface CreateContractMutationPayload extends ContractMutationPayload {
  leadId: number;
}

export function toCreateContractPayload(
  data: CreateContractRequest
): CreateContractMutationPayload {
  return {
    leadId: data.leadId,
    plan: data.plan,
    mrr: data.mrr,
    seats: data.seats,
    signedAt: data.signedAt?.trim() || undefined,
    signedByName: data.signedBy?.trim() ?? '',
    closedBy: toOptionalFiniteNumber(data.closedBy),
    notes: data.notes,
  };
}

export function toContractUpdatePayload(data: UpdateContractRequest): ContractMutationPayload {
  return {
    plan: data.plan,
    mrr: data.mrr,
    seats: data.seats,
    signedAt: data.signedAt?.trim() || undefined,
    signedByName: data.signedBy?.trim(),
    closedBy: toOptionalFiniteNumber(data.closedBy),
    notes: data.notes,
  };
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

  createCustomer: (data: Partial<Customer>) =>
    safe.post<Customer>('/crm/customers', toCustomerMutationPayload(data)),

  updateCustomer: (id: number, data: Partial<Customer>) =>
    safe.put<Customer>('/crm/customers', { id, ...toCustomerMutationPayload(data) }),

  deleteCustomer: (id: number) => safe.delete('/crm/customers', { data: { id } }),

  recordVisit: (customerId: number) => safe.post('/crm/customers/record-visit', { customerId }),

  importCustomers: (customers: Record<string, unknown>[], _dedupe?: boolean) =>
    safe.post('/crm/customers/import', { customers } satisfies ImportCustomersPayload),

  mergeCustomers: (primaryId: number, secondaryId: number) =>
    safe.post('/crm/customers/merge', { primaryId, secondaryId }),

  exportCustomerData: (id: number) => safe.get('/crm/customers/export', { params: { id } }),

  // Loyalty
  awardPoints: (customerId: number, points: number, reason?: string) =>
    safe.post('/crm/loyalty/award', { customerId, points, description: reason }),

  redeemPoints: (customerId: number, points: number, reason?: string) =>
    safe.post('/crm/loyalty/redeem', { customerId, points, description: reason }),

  getLoyaltyConfig: () => safe.get<LoyaltyConfig>('/crm/loyalty/config'),

  updateLoyaltyConfig: (config: Partial<LoyaltyConfig>) =>
    safe.put<LoyaltyConfig>('/crm/loyalty/config', toLoyaltyConfigPayload(config)),

  getLoyaltyTransactions: () => safe.get<LoyaltyTransaction[]>('/crm/loyalty/transactions'),

  getLoyaltyMembers: () => safe.get<LoyaltyMember[]>('/crm/loyalty/members'),

  getLoyaltyRewards: () => safe.get<LoyaltyReward[]>('/crm/loyalty/rewards'),

  getLoyaltyTiers: () => safe.get<LoyaltyTier[]>('/crm/loyalty/tiers'),

  updateLoyaltyTier: (id: number, data: Partial<LoyaltyTier>) =>
    safe.put('/crm/loyalty/tiers', { id, ...data }),

  // Segments
  getSegments: () => safe.get<Segment[]>('/crm/segments'),

  getSegmentById: (id: number) => safe.get<Segment>('/crm/segments/detail', { params: { id } }),

  getSegmentCustomers: (id: number) =>
    safe.get<Customer[]>('/crm/segments/customers', { params: { segment_id: id } }),

  createSegment: (data: Partial<Segment>) =>
    safe.post<Segment>('/crm/segments', toSegmentMutationPayload(data)),

  updateSegment: (id: number, data: Partial<Segment>) =>
    safe.put<Segment>('/crm/segments', { id, ...data }),

  deleteSegment: (id: number) => safe.delete('/crm/segments', { data: { id } }),

  // Campaigns
  getCampaigns: () => safe.get<Campaign[]>('/crm/campaigns'),

  getCampaignById: (id: number) => safe.get<Campaign>('/crm/campaigns/detail', { params: { id } }),

  getCampaignStats: (id: number) =>
    safe.get<CampaignStats>('/crm/campaigns/stats', { params: { campaign_id: id } }),

  createCampaign: (data: Partial<Campaign>) =>
    safe.post<Campaign>('/crm/campaigns', toCampaignMutationPayload(data)),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    safe.put<Campaign>(`/crm/campaigns/${id}`, toCampaignMutationPayload(data)),

  updateCampaignStatus: (id: number, status: string) =>
    safe.put(`/crm/campaigns/${id}/status`, { status }),

  deleteCampaign: (id: number) => safe.delete(`/crm/campaigns/${id}`),

  // Automations
  getAutomations: () => safe.get<Automation[]>('/crm/automations'),

  createAutomation: (data: Partial<Automation>) =>
    safe.post<Automation>('/crm/automations', toAutomationMutationPayload(data)),

  updateAutomation: (id: number, data: Partial<Automation>) =>
    safe.put<Automation>(`/crm/automations/${id}`, toAutomationMutationPayload(data)),

  deleteAutomation: (id: number) => safe.delete(`/crm/automations/${id}`),

  // Templates
  getTemplates: () => safe.get<Template[]>('/crm/templates'),

  createTemplate: (data: Omit<Template, 'id' | 'updatedAt'>) =>
    safe.post<Template>('/crm/templates', { ...data, type: data.channel }),

  // Feedback
  getFeedback: () => safe.get<FeedbackItem[]>('/crm/feedback'),

  // Settings
  getSettings: () => safe.get<Settings>('/crm/settings'),

  updateSettings: (data: Partial<Settings>) =>
    safe.put<Settings>('/crm/settings', toSettingsMutationPayload(data)),

  // Sales Pipeline
  getSalesPipeline: () => safe.get<PipelineColumn[]>('/crm/sales/pipeline'),
  getLeads: () => safe.get<Lead[]>('/crm/sales/leads'),
  getLead: (id: number) => safe.get<Lead>(`/crm/sales/leads/${id}`),
  createLead: (data: CreateLeadRequest) =>
    safe.post<Lead>('/crm/sales/leads', toLeadMutationPayload(data)),
  updateLead: (id: number, data: UpdateLeadRequest) =>
    safe.put<Lead>(`/crm/sales/leads/${id}`, toLeadMutationPayload(data)),
  deleteLead: (id: number) => safe.delete(`/crm/sales/leads/${id}`),
  getLeadTouchpoints: (leadId: number) =>
    safe.get<Touchpoint[]>(`/crm/sales/leads/${leadId}/touchpoints`),
  createTouchpoint: (data: CreateTouchpointRequest) =>
    safe.post<Touchpoint>('/crm/sales/touchpoints', toTouchpointMutationPayload(data)),
  deleteTouchpoint: (id: number) => safe.delete(`/crm/sales/touchpoints/${id}`),
  getContacts: (leadId?: number) =>
    safe.get<SalesContact[]>('/crm/sales/contacts', {
      params: leadId ? { lead_id: leadId } : {},
    }),
  createContact: (data: CreateContactRequest) =>
    safe.post<SalesContact>('/crm/sales/contacts', data),
  updateContact: (id: number, data: UpdateContactRequest) =>
    safe.put<SalesContact>(`/crm/sales/contacts/${id}`, data),
  deleteContact: (id: number) => safe.delete(`/crm/sales/contacts/${id}`),
  getContracts: () => safe.get<Contract[]>('/crm/sales/contracts'),
  getContract: (id: number) => safe.get<Contract>(`/crm/sales/contracts/${id}`),
  createContract: (data: CreateContractRequest) =>
    safe.post<Contract>('/crm/sales/contracts', toCreateContractPayload(data)),
  updateContract: (id: number, data: UpdateContractRequest) =>
    safe.put<Contract>(`/crm/sales/contracts/${id}`, toContractUpdatePayload(data)),
};
