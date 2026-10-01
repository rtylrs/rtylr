import { createOneKitClient, createRestRouteClient, unwrap } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CustomerServiceClient } from '@voxire/voxie/crm/customer/v1/client';
import { SegmentServiceClient } from '@voxire/voxie/crm/segment/v1/client';
import { CampaignServiceClient } from '@voxire/voxie/crm/campaign/v1/client';
import { LoyaltyServiceClient } from '@voxire/voxie/crm/loyalty/v1/client';
import { TemplateServiceClient } from '@voxire/voxie/crm/template/v1/client';
import { FeedbackServiceClient } from '@voxire/voxie/crm/feedback/v1/client';
import { SettingsServiceClient } from '@voxire/voxie/crm/settings/v1/client';
import { DashboardServiceClient } from '@voxire/voxie/crm/dashboard/v1/client';
const rest = createRestRouteClient('crm');

export const userClient = createOneKitClient(UserServiceClient, 'crm');
export const crmCustomerClient = createOneKitClient(CustomerServiceClient, 'crm');
export const crmSegmentClient = createOneKitClient(SegmentServiceClient, 'crm');

const campaignClient = createOneKitClient(CampaignServiceClient, 'crm');
const loyaltyClient = createOneKitClient(LoyaltyServiceClient, 'crm');
const templateClient = createOneKitClient(TemplateServiceClient, 'crm');
const feedbackClient = createOneKitClient(FeedbackServiceClient, 'crm');
const settingsClient = createOneKitClient(SettingsServiceClient, 'crm');
const dashboardClient = createOneKitClient(DashboardServiceClient, 'crm');

// ── REST helpers for the calls no contract declares ─────────────────────────
//
// Every path below is a literal handler in onion's router, checked rather than
// guessed. These are the endpoints crm has that were never given a .onk
// definition, which is the general problem tracked in #873.

type Req = Record<string, unknown> | undefined;

const params = (req: Req): Record<string, unknown> => {
  const { body: _body, ...rest } = (req ?? {}) as Record<string, unknown>;
  return rest;
};

const payload = (req: Req): unknown => {
  const { body: b, ...rest } = (req ?? {}) as Record<string, unknown>;
  return b ?? rest;
};

const restGet = (path: string) => (req?: Req) =>
  unwrap(rest.get<unknown>(path, { params: params(req) }));

const restPost = (path: string) => (req?: Req) => unwrap(rest.post<unknown>(path, payload(req)));

const restPut = (path: (id: string) => string) => (req: { id: string; body?: unknown }) =>
  unwrap(rest.put<unknown>(path(req.id), req.body ?? {}));

const restDelete = (path: (id: string) => string) => (req: { id: string }) =>
  unwrap(rest.delete<unknown>(path(req.id)));

const restGetById = (path: (id: string) => string) => (req: { id: string }) =>
  unwrap(rest.get<unknown>(path(req.id)));

/**
 * c6e4ce609 routed 62 of this app's calls through a single
 * `@voxire/voxie/crm/compat/v1` client. That module has never existed at any
 * voxie commit, so the app could not be type-checked or built (#894).
 *
 * Thirty of the 62 have a real contract and are dispatched to it: customer,
 * segment, campaign, loyalty, template, feedback, settings and dashboard. Where
 * the invented name disagreed with the contract, the contract wins:
 * `getCampaign` is `getCampaignDetail`, `getSegment` is `getSegmentDetail`,
 * `getLoyaltyTiers` is `listLoyaltyTiers`, and `getLoyaltyTransactions` is
 * `listLoyaltyTransactions`.
 *
 * The other 32 have no contract at all. Sales - contacts, leads, contracts,
 * touchpoints, pipeline - has none, nor do automations, the customer
 * import/export/merge/timeline endpoints, or loyalty members and rewards. Those
 * go over REST at paths verified against onion's router.
 *
 * Delegation is lazy rather than bound at module load, so a partial test mock
 * does not crash the module on import.
 */
export const crmCompatibilityClient = {
  // ── Dashboard ────────────────────────────────────────────────────────────
  getDashboard: (...args: Parameters<typeof dashboardClient.getDashboard>) =>
    dashboardClient.getDashboard(...args),

  // ── Customers ────────────────────────────────────────────────────────────
  listCustomers: (...args: Parameters<typeof crmCustomerClient.listCustomers>) =>
    crmCustomerClient.listCustomers(...args),
  getCustomer: (...args: Parameters<typeof crmCustomerClient.getCustomer>) =>
    crmCustomerClient.getCustomer(...args),
  createCustomer: (...args: Parameters<typeof crmCustomerClient.createCustomer>) =>
    crmCustomerClient.createCustomer(...args),
  updateCustomer: (...args: Parameters<typeof crmCustomerClient.updateCustomer>) =>
    crmCustomerClient.updateCustomer(...args),

  getCustomerFull: (id: number) => restGet('/crm/customers/detail/full')({ id }),
  getCustomerTimeline: restGet('/crm/customers/timeline'),
  exportCustomer: restGet('/crm/customers/export'),
  importCustomers: restPost('/crm/customers/import'),
  mergeCustomers: restPost('/crm/customers/merge'),
  recordVisit: restPost('/crm/customers/record-visit'),

  // ── Segments ─────────────────────────────────────────────────────────────
  listSegments: (...args: Parameters<typeof crmSegmentClient.listSegments>) =>
    crmSegmentClient.listSegments(...args),
  getSegment: (...args: Parameters<typeof crmSegmentClient.getSegmentDetail>) =>
    crmSegmentClient.getSegmentDetail(...args),
  getSegmentCustomers: (...args: Parameters<typeof crmSegmentClient.getSegmentCustomers>) =>
    crmSegmentClient.getSegmentCustomers(...args),
  createSegment: (...args: Parameters<typeof crmSegmentClient.createSegment>) =>
    crmSegmentClient.createSegment(...args),
  updateSegment: (...args: Parameters<typeof crmSegmentClient.updateSegment>) =>
    crmSegmentClient.updateSegment(...args),

  // ── Campaigns ────────────────────────────────────────────────────────────
  listCampaigns: (...args: Parameters<typeof campaignClient.listCampaigns>) =>
    campaignClient.listCampaigns(...args),
  getCampaign: (...args: Parameters<typeof campaignClient.getCampaignDetail>) =>
    campaignClient.getCampaignDetail(...args),
  getCampaignStats: (...args: Parameters<typeof campaignClient.getCampaignStats>) =>
    campaignClient.getCampaignStats(...args),
  createCampaign: (...args: Parameters<typeof campaignClient.createCampaign>) =>
    campaignClient.createCampaign(...args),
  updateCampaign: (...args: Parameters<typeof campaignClient.updateCampaign>) =>
    campaignClient.updateCampaign(...args),
  updateCampaignStatus: (...args: Parameters<typeof campaignClient.updateCampaignStatus>) =>
    campaignClient.updateCampaignStatus(...args),

  // campaign/v1 declares no delete.
  deleteCampaign: restDelete((id) => `/crm/campaigns/${id}`),

  // ── Loyalty ──────────────────────────────────────────────────────────────
  getLoyaltyConfig: (...args: Parameters<typeof loyaltyClient.getLoyaltyConfig>) =>
    loyaltyClient.getLoyaltyConfig(...args),
  updateLoyaltyConfig: (...args: Parameters<typeof loyaltyClient.updateLoyaltyConfig>) =>
    loyaltyClient.updateLoyaltyConfig(...args),
  getLoyaltyTiers: (...args: Parameters<typeof loyaltyClient.listLoyaltyTiers>) =>
    loyaltyClient.listLoyaltyTiers(...args),
  createLoyaltyTier: (...args: Parameters<typeof loyaltyClient.createLoyaltyTier>) =>
    loyaltyClient.createLoyaltyTier(...args),
  updateLoyaltyTier: (...args: Parameters<typeof loyaltyClient.updateLoyaltyTier>) =>
    loyaltyClient.updateLoyaltyTier(...args),
  getLoyaltyTransactions: (...args: Parameters<typeof loyaltyClient.listLoyaltyTransactions>) =>
    loyaltyClient.listLoyaltyTransactions(...args),
  awardPoints: (...args: Parameters<typeof loyaltyClient.awardPoints>) =>
    loyaltyClient.awardPoints(...args),
  redeemPoints: (...args: Parameters<typeof loyaltyClient.redeemPoints>) =>
    loyaltyClient.redeemPoints(...args),

  getLoyaltyMembers: restGet('/crm/loyalty/members'),
  getLoyaltyRewards: restGet('/crm/loyalty/rewards'),
  upsertLoyaltyReward: restPost('/crm/loyalty/rewards/upsert'),

  // ── Templates ────────────────────────────────────────────────────────────
  listTemplates: (...args: Parameters<typeof templateClient.listTemplates>) =>
    templateClient.listTemplates(...args),
  createTemplate: (...args: Parameters<typeof templateClient.createTemplate>) =>
    templateClient.createTemplate(...args),

  // ── Feedback ─────────────────────────────────────────────────────────────
  listFeedback: (...args: Parameters<typeof feedbackClient.listFeedback>) =>
    feedbackClient.listFeedback(...args),

  // ── Settings ─────────────────────────────────────────────────────────────
  getSettings: (...args: Parameters<typeof settingsClient.getSettings>) =>
    settingsClient.getSettings(...args),
  updateSettings: (...args: Parameters<typeof settingsClient.updateSettings>) =>
    settingsClient.updateSettings(...args),

  // ── Automations: no contract ─────────────────────────────────────────────
  listAutomations: restGet('/crm/automations'),
  getAutomation: async (req: { id: string }) => {
    const response = (await restGet('/crm/automations')()) as
      unknown[] | { automations?: unknown[] };
    const automations = Array.isArray(response) ? response : (response.automations ?? []);
    const automation = automations.find((item) => {
      if (!item || typeof item !== 'object') return false;
      return String((item as { id?: unknown }).id ?? '') === req.id;
    });
    if (!automation) throw new Error(`Automation ${req.id} not found`);
    return automation;
  },
  createAutomation: restPost('/crm/automations'),
  updateAutomation: restPut((id) => `/crm/automations/${id}`),
  deleteAutomation: restDelete((id) => `/crm/automations/${id}`),

  // ── Sales: no contract for any of it ─────────────────────────────────────
  listContacts: restGet('/crm/sales/contacts'),
  createContact: restPost('/crm/sales/contacts'),
  updateContact: restPut((id) => `/crm/sales/contacts/${id}`),
  deleteContact: restDelete((id) => `/crm/sales/contacts/${id}`),

  listContracts: restGet('/crm/sales/contracts'),
  getContract: restGetById((id) => `/crm/sales/contracts/${id}`),
  createContract: restPost('/crm/sales/contracts'),
  updateContract: restPut((id) => `/crm/sales/contracts/${id}`),

  listLeads: restGet('/crm/sales/leads'),
  getLead: restGetById((id) => `/crm/sales/leads/${id}`),
  createLead: restPost('/crm/sales/leads'),
  updateLead: restPut((id) => `/crm/sales/leads/${id}`),
  deleteLead: restDelete((id) => `/crm/sales/leads/${id}`),

  listTouchpoints: restGetById((id) => `/crm/sales/leads/${encodeURIComponent(id)}/touchpoints`),
  createTouchpoint: restPost('/crm/sales/touchpoints'),
  deleteTouchpoint: restDelete((id) => `/crm/sales/touchpoints/${id}`),

  getSalesPipeline: restGet('/crm/sales/pipeline'),
};
