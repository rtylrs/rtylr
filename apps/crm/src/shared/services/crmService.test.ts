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

import { crmService } from './crmService';
import { safe } from '../lib/api';

describe('crmService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getDashboard calls GET /crm/dashboard', () => {
    crmService.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/crm/dashboard');
  });

  it('getCustomers calls GET /crm/customers with params', () => {
    const params = { search: 'test' };
    crmService.getCustomers(params);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers', { params });
  });

  it('createCustomer calls POST /crm/customers', () => {
    const data = { name: 'Jane' };
    crmService.createCustomer(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers', data);
  });

  it('getCustomer calls GET /crm/customers/detail with id param', () => {
    crmService.getCustomer(3);
    expect(safe.get).toHaveBeenCalledWith('/crm/customers/detail', { params: { id: 3 } });
  });

  it('updateCustomer calls PUT /crm/customers with id', () => {
    crmService.updateCustomer(3, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/crm/customers', { id: 3, name: 'Updated' });
  });

  it('deleteCustomer calls DELETE /crm/customers with id', () => {
    crmService.deleteCustomer(3);
    expect(safe.delete).toHaveBeenCalledWith('/crm/customers', { data: { id: 3 } });
  });

  it('recordVisit calls POST /crm/customers/record-visit', () => {
    const data = { customerId: 1 };
    crmService.recordVisit(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/customers/record-visit', data);
  });

  // Loyalty
  it('getLoyaltyTransactions calls GET /crm/loyalty/transactions', () => {
    crmService.getLoyaltyTransactions();
    expect(safe.get).toHaveBeenCalledWith('/crm/loyalty/transactions', { params: undefined });
  });

  it('awardPoints calls POST /crm/loyalty/award', () => {
    const data = { customerId: 1, points: 50 };
    crmService.awardPoints(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/loyalty/award', data);
  });

  it('redeemPoints calls POST /crm/loyalty/redeem', () => {
    const data = { customerId: 1, points: 25 };
    crmService.redeemPoints(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/loyalty/redeem', data);
  });

  it('getLoyaltyConfig calls GET /crm/loyalty/config', () => {
    crmService.getLoyaltyConfig();
    expect(safe.get).toHaveBeenCalledWith('/crm/loyalty/config');
  });

  it('updateLoyaltyConfig calls PUT /crm/loyalty/config', () => {
    const data = { pointsPerDollar: 2 };
    crmService.updateLoyaltyConfig(data);
    expect(safe.put).toHaveBeenCalledWith('/crm/loyalty/config', data);
  });

  // Segments
  it('getSegments calls GET /crm/segments', () => {
    crmService.getSegments();
    expect(safe.get).toHaveBeenCalledWith('/crm/segments');
  });

  it('createSegment calls POST /crm/segments', () => {
    const data = { name: 'VIP' };
    crmService.createSegment(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/segments', data);
  });

  it('updateSegment calls PUT /crm/segments/:id', () => {
    crmService.updateSegment(1, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/crm/segments/1', { name: 'Updated' });
  });

  it('deleteSegment calls DELETE /crm/segments/:id', () => {
    crmService.deleteSegment(1);
    expect(safe.delete).toHaveBeenCalledWith('/crm/segments/1');
  });

  // Campaigns
  it('getCampaigns calls GET /crm/campaigns', () => {
    crmService.getCampaigns();
    expect(safe.get).toHaveBeenCalledWith('/crm/campaigns');
  });

  it('createCampaign calls POST /crm/campaigns', () => {
    const data = { name: 'Flash sale' };
    crmService.createCampaign(data);
    expect(safe.post).toHaveBeenCalledWith('/crm/campaigns', data);
  });

  it('updateCampaign calls PUT /crm/campaigns/:id', () => {
    crmService.updateCampaign(2, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/crm/campaigns/2', { name: 'Updated' });
  });

  it('updateCampaignStatus calls PUT /crm/campaigns/:id/status', () => {
    crmService.updateCampaignStatus(2, 'active');
    expect(safe.put).toHaveBeenCalledWith('/crm/campaigns/2/status', { status: 'active' });
  });
});
