import { describe, expect, it, vi } from 'vitest';

const mockClients = vi.hoisted(() => {
  const makeClient = () => {
    const methods: Record<string, ReturnType<typeof vi.fn>> = {};
    return new Proxy(methods, {
      get(target, property: string) {
        target[property] ??= vi.fn(async () => ({}));
        return target[property];
      },
    });
  };

  return {
    compatibility: makeClient(),
    customer: makeClient(),
    segment: makeClient(),
  };
});

vi.mock('../lib/clients', () => ({
  crmCompatibilityClient: mockClients.compatibility,
  crmCustomerClient: mockClients.customer,
  crmSegmentClient: mockClients.segment,
}));

import { crmApi, toContractUpdatePayload, toCreateContractPayload } from './crmApi';

describe('crmApi', () => {
  it('routes legacy CRM reads through the generated compatibility client', async () => {
    crmApi.getDashboard();
    crmApi.getCustomers({ search: 'John', tier: 'gold' });
    crmApi.getCustomerDetail(5);
    crmApi.getCampaigns(25, 50);
    await Promise.resolve();

    expect(mockClients.compatibility.getDashboard).toHaveBeenCalledWith({});

    // Every int64 field travels as a string. A number reaches the server as a
    // JSON number and fails to unmarshal, so passing `id: 5` was a 400 waiting
    // to happen, not just a type error.
    expect(mockClients.compatibility.listCustomers).toHaveBeenCalledWith({
      search: 'John',
      tier: 'gold',
      tag: undefined,
      page: undefined,
      pageSize: undefined,
    });
    expect(mockClients.compatibility.getCustomer).toHaveBeenCalledWith({ id: '5' });

    // ListCampaignsRequest pages by page/page_size, not limit/offset.
    expect(mockClients.compatibility.listCampaigns).toHaveBeenCalledWith({
      limit: '25',
      offset: '50',
    });
  });

  it('forwards customer mutations in the camelCase contract shape untouched', async () => {
    // The generated CreateCustomerRequest validates camelCase props itself;
    // any snake_case conversion here would fail validation before network.
    crmApi.createCustomer({
      firstName: 'Jane',
      lastName: 'Doe',
      email: 'jane@test.com',
      consent: { email: true },
    });
    crmApi.updateCustomer(7, { firstName: 'Jane', tags: ['vip'] });
    await Promise.resolve();

    expect(mockClients.compatibility.createCustomer).toHaveBeenCalledWith({
      firstName: 'Jane',
      lastName: 'Doe',
      email: 'jane@test.com',
      consent: { email: true },
    });

    // UpdateCustomerRequest is { id, body }: id binds the path segment, body
    // stays camelCase for the generated encoder.
    expect(mockClients.compatibility.updateCustomer).toHaveBeenCalledWith({
      id: '7',
      body: { firstName: 'Jane', tags: ['vip'] },
    });
  });

  it('maps campaign editor payloads onto CreateCampaignRequest/UpdateCampaignRequest', async () => {
    crmApi.createCampaign({
      name: 'Spring promo',
      type: 'email',
      subject: 'Hello',
      body: '<p>Hi</p>',
      audienceSegmentId: 12,
    } as unknown as Parameters<typeof crmApi.createCampaign>[0]);
    crmApi.updateCampaign(9, {
      name: 'SMS blast',
      channel: 'sms',
      content: 'Flash sale',
      scheduledAt: '2026-09-01T10:00:00Z',
    } as unknown as Parameters<typeof crmApi.updateCampaign>[1]);
    await Promise.resolve();

    // Legacy caller keys map onto the contract: type→channel, body→content,
    // audienceSegmentId→segmentId (int64 travels as a string).
    expect(mockClients.compatibility.createCampaign).toHaveBeenCalledWith({
      name: 'Spring promo',
      description: '',
      channel: 'email',
      segmentId: '12',
      subject: 'Hello',
      content: '<p>Hi</p>',
      scheduledAt: '',
    });

    expect(mockClients.compatibility.updateCampaign).toHaveBeenCalledWith({
      id: '9',
      body: {
        name: 'SMS blast',
        description: '',
        channel: 'sms',
        subject: '',
        content: 'Flash sale',
        scheduledAt: '2026-09-01T10:00:00Z',
      },
    });
  });

  it('passes automation payloads straight through to the REST layer', async () => {
    const request = {
      name: 'Winback',
      triggerType: 'inactivity',
      triggerConfig: { days: 60 },
      actions: [{ type: 'send_email' }],
      isActive: true,
    };
    crmApi.createAutomation(request);
    crmApi.updateAutomation(4, request);
    await Promise.resolve();

    // Automations have no contract yet; the safe client's interceptor owns
    // key casing, so the hook's canonical object must reach it unmodified
    // (the old mapper zeroed triggerType/actions and forced isActive).
    expect(mockClients.compatibility.createAutomation).toHaveBeenCalledWith(request);
    expect(mockClients.compatibility.updateAutomation).toHaveBeenCalledWith({
      id: '4',
      body: request,
    });
  });

  it('sends loyalty points mutations with int64 fields as strings', async () => {
    crmApi.awardPoints(5, 100);
    crmApi.redeemPoints(6, 50, 'Reward redemption');
    await Promise.resolve();

    // AwardPointsRequest / RedeemPointsRequest require string customerId and
    // points plus a non-empty description and referenceType.
    expect(mockClients.compatibility.awardPoints).toHaveBeenCalledWith({
      customerId: '5',
      points: '100',
      description: 'Manual adjustment',
      referenceType: 'manual',
    });
    expect(mockClients.compatibility.redeemPoints).toHaveBeenCalledWith({
      customerId: '6',
      points: '50',
      description: 'Reward redemption',
      referenceType: 'manual',
    });
  });

  it('forwards UpdateLoyaltyConfigRequest without remapping', async () => {
    const request = {
      isEnabled: true,
      pointsPerCurrencyUnit: 2,
      currencyPerPoint: 0.02,
      bronzeMin: '1',
      silverMin: '500',
      goldMin: '1500',
      platinumMin: '5000',
      bronzeDiscountPct: 1,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: '365',
    };
    crmApi.updateLoyaltyConfig(request);
    await Promise.resolve();

    // useLoyalty builds the canonical request; the old payload mapper renamed
    // it into a shape the typed client rejects.
    expect(mockClients.compatibility.updateLoyaltyConfig).toHaveBeenCalledWith(request);
  });

  it('sends UpdateSettingsRequest fields directly in camelCase', async () => {
    crmApi.updateSettings({
      defaultConsent: true,
      timezone: 'UTC',
      fromEmail: 'hello@rtylr.com',
      fromName: 'rtylr',
      smsSenderId: 'RTYLR',
    });
    await Promise.resolve();

    expect(mockClients.compatibility.updateSettings).toHaveBeenCalledWith({
      defaultConsent: true,
      timezone: 'UTC',
      fromEmail: 'hello@rtylr.com',
      fromName: 'rtylr',
      smsSenderId: 'RTYLR',
    });
  });

  it('keeps path-bound status bodies intact', async () => {
    crmApi.updateCampaignStatus(7, 'active');
    await Promise.resolve();

    // UpdateCampaignStatusRequest is { id, body }: id binds the path segment.
    expect(mockClients.compatibility.updateCampaignStatus).toHaveBeenCalledWith({
      id: '7',
      body: { status: 'active' },
    });
  });

  it('uses canonical generated delete contracts where they exist', async () => {
    crmApi.deleteCustomer(3);
    crmApi.deleteSegment(4);
    await Promise.resolve();

    expect(mockClients.customer.deleteCustomer).toHaveBeenCalledWith({ id: '3' });
    expect(mockClients.segment.deleteSegment).toHaveBeenCalledWith({ id: '4' });
  });

  it('maps contract creation fields and omits invalid optional ids', async () => {
    crmApi.createContract({
      leadId: 7,
      plan: 'pro',
      mrr: 125,
      seats: 4,
      signedAt: '2026-08-16',
      signedBy: '  Jane Doe  ',
      closedBy: '',
      notes: 'signed',
    });
    await Promise.resolve();
    // Sales has no contract at all, so createContract goes over REST at
    // POST /crm/sales/contracts and receives the payload directly.
    expect(mockClients.compatibility.createContract).toHaveBeenCalledWith({
      lead_id: 7,
      plan: 'pro',
      mrr: 125,
      seats: 4,
      signed_at: '2026-08-16',
      signed_by_name: 'Jane Doe',
      closed_by: undefined,
      notes: 'signed',
    });
    expect(
      toContractUpdatePayload({
        plan: 'scale',
        signedBy: 'Jane Doe',
        closedBy: 'not-a-user-id',
      })
    ).toEqual({
      plan: 'scale',
      mrr: undefined,
      seats: undefined,
      signedAt: undefined,
      signedByName: 'Jane Doe',
      closedBy: undefined,
      notes: undefined,
    });
    expect(
      toCreateContractPayload({
        leadId: 1,
        plan: 'free',
        mrr: 0,
        seats: 1,
        signedAt: '',
        signedBy: '',
        closedBy: '12',
      })
    ).toMatchObject({ leadId: 1, signedAt: undefined, signedByName: '', closedBy: 12 });
  });
});
