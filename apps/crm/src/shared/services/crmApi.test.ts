import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v: unknown) => v),
}));

import { crmApi } from './crmApi';
import { safe } from '../lib/api';

describe('crmApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Dashboard
  it('getDashboard calls GET /crm/dashboard', () => {
    crmApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/crm/dashboard');
  });

  // Customers
  it('getCustomers calls GET /crm/customers with params', () => {
    const params = { search: 'John', tier: 'gold' };
    crmApi.getCustomers(params);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers', { params });
  });

  it('getCustomerDetail calls GET /crm/customers/detail with id param', () => {
    crmApi.getCustomerDetail(5);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers/detail', { params: { id: 5 } });
  });

  it('getCustomerTimeline calls GET /crm/customers/timeline', () => {
    crmApi.getCustomerTimeline(5);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers/timeline', { params: { id: 5 } });
  });

  it('createCustomer calls POST /crm/customers', () => {
    const data = { name: 'Jane', email: 'jane@test.com' };
    crmApi.createCustomer(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers', data);
  });

  it('updateCustomer calls PUT /crm/customers with id', () => {
    crmApi.updateCustomer(3, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/crm/customers', { id: 3, name: 'Updated' });
  });

  it('deleteCustomer calls DELETE /crm/customers with id', () => {
    crmApi.deleteCustomer(3);
    expect(safe.delete).toHaveBeenCalledWith('/crm/customers', { data: { id: 3 } });
  });

  it('recordVisit calls POST /crm/customers/record-visit', () => {
    crmApi.recordVisit(7);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers/record-visit', { customerId: 7 });
  });

  it('importCustomers calls POST /crm/customers/import', () => {
    const customers = [{ name: 'A' }];
    crmApi.importCustomers(customers);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers/import', {
      customers,
      dedupe: undefined,
    });
  });

  it('mergeCustomers calls POST /crm/customers/merge', () => {
    crmApi.mergeCustomers(1, 2);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers/merge', {
      primaryId: 1,
      secondaryId: 2,
    });
  });

  it('exportCustomerData calls GET /crm/customers/export', () => {
    crmApi.exportCustomerData(5);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers/export', { params: { id: 5 } });
  });

  // Loyalty
  it('awardPoints calls POST /crm/loyalty/award', () => {
    crmApi.awardPoints(1, 100, 'birthday');
    expect(safe.post).toHaveBeenCalledWith('/crm/loyalty/award', {
      customerId: 1,
      points: 100,
      reason: 'birthday',
    });
  });

  it('redeemPoints calls POST /crm/loyalty/redeem', () => {
    crmApi.redeemPoints(1, 50);
    expect(safe.post).toHaveBeenCalledWith('/crm/loyalty/redeem', {
      customerId: 1,
      points: 50,
      reason: undefined,
    });
  });

  it('getLoyaltyConfig calls GET /crm/loyalty/config', () => {
    crmApi.getLoyaltyConfig();
    expect(safe.get).toHaveBeenCalledWith('/crm/loyalty/config');
  });

  it('updateLoyaltyConfig calls PUT /crm/loyalty/config', () => {
    const config = { pointsPerDollar: 2 };
    crmApi.updateLoyaltyConfig(config);
    expect(safe.put).toHaveBeenCalledWith('/crm/loyalty/config', config);
  });

  it('getLoyaltyMembers calls GET /crm/loyalty/members', () => {
    crmApi.getLoyaltyMembers();
    expect(safe.get).toHaveBeenCalledWith('/crm/loyalty/members');
  });

  it('updateLoyaltyTier calls PUT /crm/loyalty/tiers', () => {
    crmApi.updateLoyaltyTier(1, { name: 'Platinum' });
    expect(safe.put).toHaveBeenCalledWith('/crm/loyalty/tiers', { id: 1, name: 'Platinum' });
  });

  // Segments
  it('getSegments calls GET /crm/segments', () => {
    crmApi.getSegments();
    expect(safe.get).toHaveBeenCalledWith('/crm/segments');
  });

  it('getSegmentById calls GET /crm/segments/detail', () => {
    crmApi.getSegmentById(2);
    expect(safe.get).toHaveBeenCalledWith('/crm/segments/detail', { params: { id: 2 } });
  });

  it('createSegment calls POST /crm/segments', () => {
    const data = { name: 'VIPs' };
    crmApi.createSegment(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/segments', data);
  });

  it('deleteSegment calls DELETE /crm/segments with id', () => {
    crmApi.deleteSegment(4);
    expect(safe.delete).toHaveBeenCalledWith('/crm/segments', { data: { id: 4 } });
  });

  // Campaigns
  it('getCampaigns calls GET /crm/campaigns', () => {
    crmApi.getCampaigns();
    expect(safe.get).toHaveBeenCalledWith('/crm/campaigns');
  });

  it('createCampaign calls POST /crm/campaigns', () => {
    const data = { name: 'Summer sale' };
    crmApi.createCampaign(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/campaigns', data);
  });

  it('updateCampaignStatus calls PUT /crm/campaigns/status', () => {
    crmApi.updateCampaignStatus(1, 'active');
    expect(safe.put).toHaveBeenCalledWith('/crm/campaigns/1/status', { status: 'active' });
  });

  // Automations
  it('getAutomations calls GET /crm/automations', () => {
    crmApi.getAutomations();
    expect(safe.get).toHaveBeenCalledWith('/crm/automations');
  });

  it('createAutomation calls POST /crm/automations', () => {
    const data = { name: 'Welcome flow' };
    crmApi.createAutomation(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/automations', data);
  });

  // Templates
  it('getTemplates calls GET /crm/templates', () => {
    crmApi.getTemplates();
    expect(safe.get).toHaveBeenCalledWith('/crm/templates');
  });

  // Feedback
  it('getFeedback calls GET /crm/feedback', () => {
    crmApi.getFeedback();
    expect(safe.get).toHaveBeenCalledWith('/crm/feedback');
  });

  // Settings
  it('getSettings calls GET /crm/settings', () => {
    crmApi.getSettings();
    expect(safe.get).toHaveBeenCalledWith('/crm/settings');
  });

  it('updateSettings calls PUT /crm/settings', () => {
    const data = { fromEmail: 'hi@test.com' };
    crmApi.updateSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/crm/settings', data);
  });
});
