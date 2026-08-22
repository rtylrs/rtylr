import { fromOneKitJson, oneKitResult, toOneKitJson, toOptionalFiniteNumber } from '@voxire/api';
import type { UpdateLoyaltyConfigRequest } from '@voxire/voxie/crm/loyalty/v1/types';
import { crmCompatibilityClient, crmCustomerClient, crmSegmentClient } from '../lib/clients';

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
  /** Backend persists an arbitrary JSON consent record; read keys defensively. */
  consent?: Record<string, boolean>;
  createdAt?: string;
  referredBy?: number | null;
}

export interface CustomersParams {
  search?: string;
  tier?: string;
  tag?: string;
  page?: number;
  pageSize?: number;
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

/**
 * Wire input for CreateCustomerRequest / UpdateCustomerBody
 * (gen/ts crm/customer/v1/types.ts). The generated client validates camelCase
 * props and encodes snake_case itself, so callers forward this shape as-is.
 */
interface CustomerMutationPayload {
  firstName: string;
  lastName?: string;
  email?: string;
  phone?: string;
  dateOfBirth?: string;
  gender?: string;
  address?: string;
  city?: string;
  country?: string;
  tags?: string[];
  notes?: string;
  consent?: Record<string, boolean>;
}

/** UpdateCustomerBody additionally carries isActive. */
interface CustomerUpdateBody extends CustomerMutationPayload {
  isActive?: boolean;
}

interface ImportCustomersPayload {
  customers: Record<string, unknown>[];
}

interface LoyaltyPointsPayload {
  customerId: string;
  points: string;
  description: string;
  referenceId?: string;
  referenceType: string;
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

/**
 * Wire input mirroring CreateCampaignRequest / UpdateCampaignBody
 * (gen/ts crm/campaign/v1/types.ts): camelCase, int64 segmentId as string.
 */
interface CampaignMutationPayload {
  name: string;
  description: string;
  channel: string;
  segmentId?: string;
  subject: string;
  content: string;
  scheduledAt: string;
}

/**
 * Maps caller payloads onto the generated request shape. Legacy editor keys
 * are accepted alongside the contract keys: type→channel, body→content,
 * audienceSegmentId→segmentId, scheduledFor→scheduledAt. The int64 segment id
 * travels as a string on the wire.
 */
function toCampaignMutationPayload(
  data: Partial<Campaign> & {
    /** Contract key accepted alongside the legacy `body`. */
    content?: string;
    /** Contract key accepted alongside the legacy `scheduledFor`. */
    scheduledAt?: string;
    /** Legacy caller key for channel ('email' | 'sms'). */
    type?: string;
    /** Legacy caller key for segmentId. */
    audienceSegmentId?: number | null;
  }
): CampaignMutationPayload {
  const rawSegmentId = data.segmentId ?? data.audienceSegmentId;
  return {
    name: data.name ?? '',
    description: data.description ?? '',
    channel: data.channel ?? data.type ?? '',
    ...(rawSegmentId === undefined || rawSegmentId === null
      ? {}
      : { segmentId: String(rawSegmentId) }),
    subject: data.subject ?? '',
    content: data.content ?? data.body ?? '',
    scheduledAt: data.scheduledAt ?? data.scheduledFor ?? '',
  };
}

/**
 * Canonical backend shape emitted by useAutomations' toBackendRequest
 * (name, triggerType, triggerConfig, actions, isActive). Automations have no
 * .onk contract yet — these calls go over REST through the safe client, whose
 * interceptor performs the camelCase→snake_case conversion. Declared as a
 * type alias (not an interface) so it satisfies the REST layer's
 * Record<string, unknown> payload parameter via implicit index signature.
 */
type AutomationBackendRequest = {
  name: string;
  triggerType?: string;
  triggerConfig?: Record<string, unknown>;
  actions?: unknown[];
  isActive?: boolean;
};

// UpdateSettingsRequest (gen/ts crm/settings/v1/types.ts) declares its fields
// directly; the generated client validates camelCase props and encodes
// snake_case itself, so no local mapper is needed.

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

function compatResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((value) => fromOneKitJson<T>(value));
}

/**
 * Payload for a request whose fields are declared directly, with no body member.
 *
 * Both of these assert rather than check: toOneKitJson only renames keys to the
 * wire's snake_case, so nothing here could validate a shape against the request
 * type. The single loose `body()` they replace compiled against the invented
 * compat client because its parameters accepted anything.
 */
function flat<T>(value: unknown): T {
  return toOneKitJson(value) as T;
}

/** Payload for a request shaped { id, body }, where id binds a path segment. */
function withId<B>(id: string | number, value: unknown): { id: string; body: B } {
  return { id: String(id), body: toOneKitJson(value) as B };
}

/**
 * Same envelope as withId, but the body is forwarded untouched for generated
 * clients that encode camelCase props to snake_case themselves. Running
 * toOneKitJson over those bodies renames keys before validation and drops
 * every field — the client would throw TypeError before any network traffic.
 */
function typedRequest<B>(id: string | number, body: B): { id: string; body: B } {
  return { id: String(id), body };
}

/**
 * Coerce wire values into finite numbers. Money fields (mrr, mrr_potential,
 * total_mrr) are serialized by the backend's money.Amount as decimal strings
 * ("1500.00"); without this, arithmetic like reduce(sum + mrr) concatenates
 * strings. Returns 0 for anything that cannot be parsed.
 */
export function toFiniteNumber(value: unknown): number {
  if (typeof value === 'number') return Number.isFinite(value) ? value : 0;
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export const crmApi = {
  // Dashboard
  getDashboard: () => compatResult<DashboardStats>(() => crmCompatibilityClient.getDashboard({})),

  // Customers
  // ListCustomersRequest takes search, tier, tag, page and page_size, and the
  // int64 fields travel as strings. It has no limit/offset, so those two
  // CustomersParams fields were never reaching the server and are dropped here
  // rather than silently passed: see #913.
  getCustomers: (params?: CustomersParams) =>
    compatResult<Customer[]>(() =>
      crmCompatibilityClient.listCustomers({
        search: params?.search,
        tier: params?.tier,
        tag: params?.tag,
        page: params?.page === undefined ? undefined : String(params.page),
        pageSize: params?.pageSize === undefined ? undefined : String(params.pageSize),
      })
    ),

  getCustomerDetail: (id: number) =>
    compatResult<Customer>(() => crmCompatibilityClient.getCustomer({ id: String(id) })),

  getCustomerFull: (id: number) =>
    compatResult<Customer>(() => crmCompatibilityClient.getCustomerFull({ id })),

  getCustomerTimeline: (
    id: number,
    options?: { customerId?: number; cursor?: string; limit?: number }
  ) =>
    compatResult<TimelineEntry[]>(() =>
      crmCompatibilityClient.getCustomerTimeline({
        id,
        customerId: options?.customerId,
        cursor: options?.cursor,
        limit: options?.limit,
      })
    ),

  createCustomer: (data: CustomerMutationPayload) =>
    compatResult<Customer>(() => crmCompatibilityClient.createCustomer(data)),

  updateCustomer: (id: number, data: CustomerUpdateBody) =>
    compatResult<Customer>(() => crmCompatibilityClient.updateCustomer(typedRequest(id, data))),

  deleteCustomer: (id: number) =>
    oneKitResult(() => crmCustomerClient.deleteCustomer({ id: String(id) })),

  recordVisit: (customerId: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.recordVisit(flat({ customerId }))),

  importCustomers: (customers: Record<string, unknown>[], _dedupe?: boolean) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.importCustomers(flat({ customers } satisfies ImportCustomersPayload))
    ),

  mergeCustomers: (primaryId: number, secondaryId: number) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.mergeCustomers(flat({ primaryId, secondaryId }))
    ),

  exportCustomerData: (id: number, useCustomerId = false) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.exportCustomer(useCustomerId ? { customerId: id } : { id })
    ),

  // Loyalty
  // AwardPointsRequest / RedeemPointsRequest (gen/ts crm/loyalty/v1/types.ts):
  // customerId and points are int64 fields travelling as strings, description
  // and referenceType are required. The generated client validates camelCase
  // props before encoding.
  awardPoints: (customerId: number, points: number, reason?: string) =>
    compatResult<unknown>(() => {
      const payload: LoyaltyPointsPayload = {
        customerId: String(customerId),
        points: String(points),
        description: reason?.trim() || 'Manual adjustment',
        referenceType: 'manual',
      };
      return crmCompatibilityClient.awardPoints(payload);
    }),

  redeemPoints: (customerId: number, points: number, reason?: string) =>
    compatResult<unknown>(() => {
      const payload: LoyaltyPointsPayload = {
        customerId: String(customerId),
        points: String(points),
        description: reason?.trim() || 'Manual adjustment',
        referenceType: 'manual',
      };
      return crmCompatibilityClient.redeemPoints(payload);
    }),

  getLoyaltyConfig: () =>
    compatResult<LoyaltyConfig>(() => crmCompatibilityClient.getLoyaltyConfig({})),

  updateLoyaltyConfig: (config: UpdateLoyaltyConfigRequest) =>
    compatResult<LoyaltyConfig>(() => crmCompatibilityClient.updateLoyaltyConfig(config)),

  getLoyaltyTransactions: (customerId?: number, limit?: number, offset?: number) =>
    compatResult<LoyaltyTransaction[]>(() =>
      // The contract pages by page/page_size. Translating rather than dropping
      // the caller's limit/offset, which is a faithful mapping while both are
      // multiples: page is 1-based, so offset 0 is page 1.
      crmCompatibilityClient.getLoyaltyTransactions({
        customerId: customerId === undefined ? undefined : String(customerId),
        pageSize: limit === undefined ? undefined : String(limit),
        page:
          offset === undefined || limit === undefined || limit === 0
            ? undefined
            : String(Math.floor(offset / limit) + 1),
      })
    ),

  getLoyaltyMembers: (limit?: number, offset?: number) =>
    compatResult<LoyaltyMember[]>(() =>
      crmCompatibilityClient.getLoyaltyMembers({ limit, offset })
    ),

  getLoyaltyRewards: () =>
    compatResult<LoyaltyReward[]>(() => crmCompatibilityClient.getLoyaltyRewards({})),

  getLoyaltyTiers: () =>
    compatResult<LoyaltyTier[]>(() => crmCompatibilityClient.getLoyaltyTiers({})),

  updateLoyaltyTier: (id: number, data: Partial<LoyaltyTier>) =>
    compatResult<unknown>(() => crmCompatibilityClient.updateLoyaltyTier(withId(id, data))),

  createLoyaltyTier: (data: Partial<LoyaltyTier>) =>
    compatResult<unknown>(() => crmCompatibilityClient.createLoyaltyTier(flat(data))),

  upsertLoyaltyReward: (id: number | undefined, data: Partial<LoyaltyReward>) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.upsertLoyaltyReward(flat(id ? { id, ...data } : data))
    ),

  // Segments
  getSegments: () => compatResult<Segment[]>(() => crmCompatibilityClient.listSegments({})),

  getSegmentById: (id: number) =>
    compatResult<Segment>(() => crmCompatibilityClient.getSegment({ id: String(id) })),

  createSegment: (data: Partial<Segment>) =>
    compatResult<Segment>(() =>
      crmCompatibilityClient.createSegment(flat(toSegmentMutationPayload(data)))
    ),

  updateSegment: (id: number, data: Partial<Segment>) =>
    compatResult<Segment>(() => crmCompatibilityClient.updateSegment(withId(id, data))),

  deleteSegment: (id: number) =>
    oneKitResult(() => crmSegmentClient.deleteSegment({ id: String(id) })),

  // Campaigns
  getCampaigns: (limit?: number, offset?: number) =>
    compatResult<Campaign[]>(() =>
      crmCompatibilityClient.listCampaigns({
        limit: limit === undefined ? undefined : String(limit),
        offset: offset === undefined ? undefined : String(offset),
      })
    ),

  getCampaignById: (id: number) =>
    compatResult<Campaign>(() => crmCompatibilityClient.getCampaign({ id: String(id) })),

  getCampaignStats: (id: number) =>
    compatResult<CampaignStats>(() =>
      crmCompatibilityClient.getCampaignStats({ campaignId: String(id) })
    ),

  createCampaign: (data: Partial<Campaign>) =>
    compatResult<Campaign>(() =>
      crmCompatibilityClient.createCampaign(toCampaignMutationPayload(data))
    ),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    compatResult<Campaign>(() =>
      crmCompatibilityClient.updateCampaign(typedRequest(id, toCampaignMutationPayload(data)))
    ),

  updateCampaignStatus: (id: number, status: string) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.updateCampaignStatus(withId(id, { status }))
    ),

  deleteCampaign: (id: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.deleteCampaign({ id: String(id) })),

  // Automations
  getAutomations: () =>
    compatResult<Automation[]>(() => crmCompatibilityClient.listAutomations({})),

  getAutomation: (id: number) =>
    compatResult<Automation>(() => crmCompatibilityClient.getAutomation({ id: String(id) })),

  createAutomation: (data: AutomationBackendRequest) =>
    compatResult<Automation>(() => crmCompatibilityClient.createAutomation(data)),

  updateAutomation: (id: number, data: AutomationBackendRequest) =>
    compatResult<Automation>(() => crmCompatibilityClient.updateAutomation(typedRequest(id, data))),

  deleteAutomation: (id: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.deleteAutomation({ id: String(id) })),

  // Templates
  getTemplates: () => compatResult<Template[]>(() => crmCompatibilityClient.listTemplates({})),

  createTemplate: (data: Omit<Template, 'id' | 'updatedAt'>) =>
    compatResult<Template>(() =>
      crmCompatibilityClient.createTemplate(flat({ ...data, type: data.channel }))
    ),

  // Feedback
  getFeedback: () => compatResult<FeedbackItem[]>(() => crmCompatibilityClient.listFeedback({})),

  // Settings
  getSettings: () => compatResult<Settings>(() => crmCompatibilityClient.getSettings({})),

  updateSettings: (data: Partial<Settings>) =>
    compatResult<Settings>(() =>
      // UpdateSettingsRequest declares its fields directly; the generated
      // client validates camelCase props and encodes snake_case itself.
      crmCompatibilityClient.updateSettings({
        defaultConsent: data.defaultConsent,
        timezone: data.timezone,
        fromEmail: data.fromEmail,
        fromName: data.fromName,
        smsSenderId: data.smsSenderId,
      })
    ),

  // Sales Pipeline
  getSalesPipeline: () =>
    compatResult<PipelineColumn[]>(() => crmCompatibilityClient.getSalesPipeline({})),
  getLeads: (limit?: number, offset?: number) =>
    compatResult<Lead[]>(() => crmCompatibilityClient.listLeads({ limit, offset })),
  getLead: (id: number) =>
    compatResult<Lead>(() => crmCompatibilityClient.getLead({ id: String(id) })),
  createLead: (data: CreateLeadRequest) =>
    compatResult<Lead>(() => crmCompatibilityClient.createLead(flat(toLeadMutationPayload(data)))),
  updateLead: (id: number, data: UpdateLeadRequest) =>
    compatResult<Lead>(() =>
      crmCompatibilityClient.updateLead(withId(id, toLeadMutationPayload(data)))
    ),
  deleteLead: (id: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.deleteLead({ id: String(id) })),
  getLeadTouchpoints: (leadId: number) =>
    compatResult<Touchpoint[]>(() =>
      crmCompatibilityClient.listTouchpoints({ id: String(leadId) })
    ),
  createTouchpoint: (data: CreateTouchpointRequest) =>
    compatResult<Touchpoint>(() =>
      crmCompatibilityClient.createTouchpoint(flat(toTouchpointMutationPayload(data)))
    ),
  deleteTouchpoint: (id: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.deleteTouchpoint({ id: String(id) })),
  getContacts: (leadId?: number, limit = 100, offset = 0) =>
    compatResult<SalesContact[]>(() =>
      crmCompatibilityClient.listContacts({ leadId, limit, offset })
    ),
  createContact: (data: CreateContactRequest) =>
    compatResult<SalesContact>(() => crmCompatibilityClient.createContact(flat(data))),
  updateContact: (id: number, data: UpdateContactRequest) =>
    compatResult<SalesContact>(() => crmCompatibilityClient.updateContact(withId(id, data))),
  deleteContact: (id: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.deleteContact({ id: String(id) })),
  getContracts: (limit?: number, offset?: number) =>
    compatResult<Contract[]>(() => crmCompatibilityClient.listContracts({ limit, offset })),
  getContract: (id: number) =>
    compatResult<Contract>(() => crmCompatibilityClient.getContract({ id: String(id) })),
  createContract: (data: CreateContractRequest) =>
    compatResult<Contract>(() =>
      crmCompatibilityClient.createContract(flat(toCreateContractPayload(data)))
    ),
  updateContract: (id: number, data: UpdateContractRequest) =>
    compatResult<Contract>(() =>
      crmCompatibilityClient.updateContract(withId(id, toContractUpdatePayload(data)))
    ),
};
