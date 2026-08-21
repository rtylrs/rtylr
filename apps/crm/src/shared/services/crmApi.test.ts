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
    expect(mockClients.compatibility.listCustomers).toHaveBeenCalledWith({
      search: 'John',
      tier: 'gold',
    });
    expect(mockClients.compatibility.getCustomer).toHaveBeenCalledWith({ id: 5 });
    expect(mockClients.compatibility.listCampaigns).toHaveBeenCalledWith({ limit: 25, offset: 50 });
  });

  it('converts compatibility request bodies to the backend wire shape', async () => {
    crmApi.createCustomer({ name: 'Jane Doe', email: 'jane@test.com' });
    crmApi.updateCampaignStatus(7, 'active');
    await Promise.resolve();

    expect(mockClients.compatibility.createCustomer).toHaveBeenCalledWith({
      body: {
        first_name: 'Jane',
        last_name: 'Doe',
        email: 'jane@test.com',
        phone: undefined,
        date_of_birth: undefined,
        tags: undefined,
        notes: undefined,
      },
    });
    expect(mockClients.compatibility.updateCampaignStatus).toHaveBeenCalledWith({
      id: '7',
      body: { id: '7', status: 'active' },
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
    expect(mockClients.compatibility.createContract).toHaveBeenCalledWith({
      body: {
        lead_id: 7,
        plan: 'pro',
        mrr: 125,
        seats: 4,
        signed_at: '2026-08-16',
        signed_by_name: 'Jane Doe',
        closed_by: undefined,
        notes: 'signed',
      },
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
