import { fromOneKitJson, oneKitResult, toOneKitJson, toOptionalFiniteNumber } from '@voxire/api';
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
  /** One-based page number, matching ListCustomersRequest.page. */
  page?: number;
  /** Page size, matching ListCustomersRequest.page_size. */
  pageSize?: number;
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
  meta?: Record<string, unknown>;
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
  /** Canonical backend fields, sent by the automations hook as the mutation payload. */
  triggerType?: string;
  triggerConfig?: Record<string, unknown> | string;
  actions?: unknown[];
  isActive?: boolean;
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
  'cold' | 'contacted' | 'demo_booked' | 'demo_done' | 'proposal_sent' | 'contracted' | 'lost';

export type TouchpointType = 'phone' | 'email' | 'meeting';
export type TouchpointOutcome = 'positive' | 'neutral' | 'negative' | 'no_answer';
/** How often a contract's value is charged. */
export type ContractBilling = 'one_time' | 'monthly' | 'yearly';

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
  // YYYY-MM-DD follow-up set on the latest touchpoint; absent when none.
  nextFollowUp?: string;
}

/** A teammate a lead can be assigned to. */
export interface SalesRep {
  id: number;
  name: string;
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
  // The CRM customer this contact was captured into, once it has been.
  customerId?: number | null;
  createdAt: string;
}

/**
 * What a won lead bought. `plan` is free text for what was sold; `value` is
 * charged once or per `billing` period over the optional term. `mrr` is the
 * server's monthly equivalent of value, kept for pipeline totals.
 */
export interface Contract {
  id: number;
  leadId: number;
  company: string;
  plan: string;
  mrr: number;
  seats: number;
  value: number;
  billing: ContractBilling;
  startDate?: string;
  endDate?: string;
  customerId?: number | null;
  // YYYY-MM-DD; absent while the contract is an unsigned draft.
  signedAt: string;
  signedBy: string;
  closedBy: number | null;
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
  // null unassigns the lead; undefined keeps the current rep.
  assignedTo?: number | null;
  notes?: string;
}

interface LeadMutationPayload {
  companyName: string;
  industry?: string;
  size?: string;
  source?: string;
  stage?: LeadStage;
  assignedTo?: number | null;
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
  gender?: string;
  address?: string;
  city?: string;
  country?: string;
  tags?: string[];
  notes?: string;
  isActive?: boolean;
  consent?: Record<string, boolean>;
}

type CustomerMutationInput = Partial<Customer> & {
  firstName?: string;
  lastName?: string;
  dateOfBirth?: string;
  gender?: string;
  address?: string;
  city?: string;
  country?: string;
  isActive?: boolean;
};

function splitCustomerName(name?: string): { firstName: string; lastName: string } {
  const parts = (name ?? '').trim().split(/\s+/).filter(Boolean);
  return {
    firstName: parts[0] ?? '',
    lastName: parts.slice(1).join(' '),
  };
}

function toCustomerMutationPayload(data: CustomerMutationInput): CustomerMutationPayload {
  const firstName = data.firstName?.trim() ?? '';
  const lastName = data.lastName?.trim() ?? '';
  const names = firstName || lastName ? { firstName, lastName } : splitCustomerName(data.name);
  // useCustomers' wire input emits `dateOfBirth` (matching the generated
  // contract); `birthday` is the schema's display alias. Reading only
  // `birthday` dropped the birthday, address and consent the create form
  // collects (B50-R04).
  return {
    firstName: names.firstName,
    lastName: names.lastName,
    email: data.email,
    phone: data.phone,
    dateOfBirth: data.dateOfBirth ?? data.birthday,
    gender: data.gender,
    address: data.address,
    city: data.city,
    country: data.country,
    tags: data.tags,
    notes: data.notes,
    isActive: data.isActive,
    consent: data.consent,
  };
}

interface ImportCustomersPayload {
  customers: Record<string, unknown>[];
  dedupe?: boolean;
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
  description?: string;
  channel?: string;
  segmentId?: string;
  subject?: string;
  content?: string;
  scheduledAt?: string;
}

function toCampaignMutationPayload(
  data: Partial<Campaign> & Record<string, unknown>
): CampaignMutationPayload {
  // The editors send `type`, never `channel`. Reading only `channel` meant the
  // fallback always won, so every campaign was saved as email whatever the
  // operator picked.
  const channel = (data.type as string | undefined) ?? data.channel ?? 'email';

  // Nothing is invented any more. description and subject used to fall back to
  // the name, and scheduledAt was stamped with the current time on every save,
  // because the contract marked all three @required. voxie#73 relaxed them to
  // match what onion validates (name, channel, content), so a value the operator
  // did not supply is now simply absent instead of fabricated. Stamping
  // scheduledAt was the worst of the three: it asserts a draft is scheduled for
  // this instant, which anything reading scheduled_at could act on.
  const description = typeof data.description === 'string' ? data.description : undefined;
  const subject = typeof data.subject === 'string' ? data.subject : undefined;
  const scheduledAt = typeof data.scheduledAt === 'string' ? data.scheduledAt : undefined;

  return {
    name: data.name,
    channel,
    segmentId: data.audienceSegmentId != null ? String(data.audienceSegmentId) : undefined,
    content: data.body ?? '',
    ...(description ? { description } : {}),
    ...(subject ? { subject } : {}),
    ...(scheduledAt ? { scheduledAt } : {}),
  };
}

interface AutomationMutationPayload {
  name?: string;
  triggerType?: string;
  triggerConfig?: Record<string, unknown>;
  actions?: unknown[];
  isActive?: boolean;
}

/**
 * The automations hook already converts its legacy UI record into the backend
 * shape (triggerType/triggerConfig/actions/isActive). Reading only the legacy
 * fields here dropped trigger_type and actions from every PUT, which onion
 * rejects with 400. Prefer the canonical fields and keep the legacy ones as a
 * fallback for callers that still send the UI record.
 */
function toAutomationMutationPayload(data: Partial<Automation>): AutomationMutationPayload {
  const triggerConfig =
    data.triggerConfig && typeof data.triggerConfig === 'object' ? data.triggerConfig : {};
  return {
    name: data.name,
    triggerType: data.triggerType ?? data.trigger,
    triggerConfig,
    actions: data.actions ?? data.nodes,
    isActive: data.isActive ?? data.status === 'active',
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
  plan: string;
  value: number;
  billing: ContractBilling;
  startDate?: string;
  endDate?: string;
  signedAt?: string;
  signedBy?: string;
  closedBy?: number | null;
  notes?: string;
}

export type UpdateContractRequest = Partial<Omit<CreateContractRequest, 'leadId'>>;

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
  plan?: string;
  value?: number;
  billing?: ContractBilling;
  startDate?: string;
  endDate?: string;
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
    plan: data.plan.trim(),
    value: data.value,
    billing: data.billing,
    startDate: data.startDate?.trim() || undefined,
    endDate: data.endDate?.trim() || undefined,
    signedAt: data.signedAt?.trim() || undefined,
    signedByName: data.signedBy?.trim() ?? '',
    closedBy: toOptionalFiniteNumber(data.closedBy),
    notes: data.notes,
  };
}

export function toContractUpdatePayload(data: UpdateContractRequest): ContractMutationPayload {
  return {
    plan: data.plan?.trim(),
    value: data.value,
    billing: data.billing,
    startDate: data.startDate?.trim() || undefined,
    endDate: data.endDate?.trim() || undefined,
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

/** Payload for a generated request with a path ID and typed camelCase body. */
function withCanonicalId<B>(id: string | number, body: B): { id: string; body: B } {
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
  // int64 fields travel as strings (#913 removed the phantom limit/offset).
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
    compatResult<Customer>(() => crmCompatibilityClient.getCustomerFull(id)),

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

  createCustomer: (data: CustomerMutationInput) =>
    compatResult<Customer>(() =>
      crmCompatibilityClient.createCustomer(toCustomerMutationPayload(data))
    ),

  updateCustomer: (id: number, data: Partial<Customer>) =>
    compatResult<Customer>(() =>
      crmCompatibilityClient.updateCustomer(withCanonicalId(id, toCustomerMutationPayload(data)))
    ),

  deleteCustomer: (id: number) =>
    oneKitResult(() => crmCustomerClient.deleteCustomer({ id: String(id) })),

  recordVisit: (customerId: number) =>
    compatResult<unknown>(() => crmCompatibilityClient.recordVisit(flat({ customerId }))),

  importCustomers: (customers: Record<string, unknown>[], dedupe?: boolean) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.importCustomers(
        flat({ customers, dedupe } satisfies ImportCustomersPayload)
      )
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
  awardPoints: (customerId: number, points: number, reason?: string) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.awardPoints(
        flat({
          customerId: String(customerId),
          points: String(points),
          description: reason && reason.trim() !== '' ? reason : 'Manual adjustment',
          referenceType: 'manual',
        })
      )
    ),

  redeemPoints: (customerId: number, points: number, reason?: string) =>
    compatResult<unknown>(() =>
      crmCompatibilityClient.redeemPoints(
        flat({
          customerId: String(customerId),
          points: String(points),
          description: reason && reason.trim() !== '' ? reason : 'Manual adjustment',
          referenceType: 'manual',
        })
      )
    ),

  getLoyaltyConfig: () =>
    compatResult<LoyaltyConfig>(() => crmCompatibilityClient.getLoyaltyConfig({})),

  updateLoyaltyConfig: (config: Partial<LoyaltyConfig>) =>
    compatResult<LoyaltyConfig>(() => crmCompatibilityClient.updateLoyaltyConfig(flat(config))),

  getLoyaltyTransactions: (customerId?: number, page?: number, pageSize?: number) =>
    compatResult<LoyaltyTransaction[]>(() =>
      // Contract paging verbatim: ListLoyaltyTransactionsRequest is
      // customer_id, page, page_size. int64 travels as a string (#913).
      crmCompatibilityClient.getLoyaltyTransactions({
        customerId: customerId === undefined ? undefined : String(customerId),
        page: page === undefined ? undefined : String(page),
        pageSize: pageSize === undefined ? undefined : String(pageSize),
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
      crmCompatibilityClient.createCampaign(flat(toCampaignMutationPayload(data)))
    ),

  updateCampaign: (id: number, data: Partial<Campaign>) =>
    compatResult<Campaign>(() =>
      crmCompatibilityClient.updateCampaign(withId(id, toCampaignMutationPayload(data)))
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

  createAutomation: (data: Partial<Automation>) =>
    compatResult<Automation>(() =>
      crmCompatibilityClient.createAutomation(flat(toAutomationMutationPayload(data)))
    ),

  updateAutomation: (id: number, data: Partial<Automation>) =>
    compatResult<Automation>(() =>
      crmCompatibilityClient.updateAutomation(withId(id, toAutomationMutationPayload(data)))
    ),

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
      crmCompatibilityClient.updateSettings(flat(toSettingsMutationPayload(data)))
    ),

  // Sales Pipeline
  getSalesPipeline: () =>
    compatResult<PipelineColumn[]>(() => crmCompatibilityClient.getSalesPipeline({})),
  getSalesReps: () => compatResult<SalesRep[]>(() => crmCompatibilityClient.listSalesReps({})),
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
