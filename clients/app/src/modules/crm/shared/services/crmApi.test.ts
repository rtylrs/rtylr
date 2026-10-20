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

import { crmApi } from './crmApi';

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

  it('passes canonical request bodies to generated clients', async () => {
    crmApi.createCustomer({ name: 'Jane Doe', email: 'jane@test.com' });
    crmApi.updateCampaignStatus(7, 'active');
    await Promise.resolve();

    // CreateCustomerRequest declares its fields directly. There is no `body`
    // member to wrap them in, so the wrapper was an extra level the server
    // never asked for.
    expect(mockClients.compatibility.createCustomer).toHaveBeenCalledWith({
      firstName: 'Jane',
      lastName: 'Doe',
      email: 'jane@test.com',
      phone: undefined,
      dateOfBirth: undefined,
      tags: undefined,
      notes: undefined,
    });

    // UpdateCampaignStatusRequest is { id, body }: id binds the path segment.
    // It used to appear in both places, which is harmless but says the shape
    // was not being read from the contract.
    expect(mockClients.compatibility.updateCampaignStatus).toHaveBeenCalledWith({
      id: '7',
      body: { status: 'active' },
    });
  });

  it('carries the birthday, address and consent the customer form collects', async () => {
    vi.clearAllMocks();
    // useCustomers' wire input emits `dateOfBirth` (the generated contract's
    // name); `birthday` is only the schema's display alias. The mapper used to
    // read only `birthday`, so the whole second step of the create form was
    // dropped on create and update (B50-R04).
    const formInput = {
      firstName: 'Jane',
      lastName: 'Doe',
      email: 'jane@test.com',
      dateOfBirth: '1990-01-01',
      gender: 'female',
      address: '12 Main St',
      city: 'Austin',
      country: 'US',
      consent: { email: true, sms: false, marketing: false },
    };

    await crmApi.createCustomer(formInput);
    expect(mockClients.compatibility.createCustomer).toHaveBeenCalledWith(
      expect.objectContaining({
        dateOfBirth: '1990-01-01',
        gender: 'female',
        address: '12 Main St',
        city: 'Austin',
        country: 'US',
        consent: { email: true, sms: false, marketing: false },
      })
    );

    await crmApi.updateCustomer(5, formInput);
    expect(mockClients.compatibility.updateCustomer).toHaveBeenCalledWith({
      id: '5',
      body: expect.objectContaining({
        dateOfBirth: '1990-01-01',
        address: '12 Main St',
        consent: { email: true, sms: false, marketing: false },
      }),
    });
  });

  it('uses canonical generated delete contracts where they exist', async () => {
    crmApi.deleteCustomer(3);
    crmApi.deleteSegment(4);
    await Promise.resolve();

    expect(mockClients.customer.deleteCustomer).toHaveBeenCalledWith({ id: '3' });
    expect(mockClients.segment.deleteSegment).toHaveBeenCalledWith({ id: '4' });
  });
});

/**
 * Campaign create and update payloads. rtylr#938.
 *
 * The editors send `{name, type, subject?, body, audienceSegmentId}`. The mapping
 * read `data.channel`, which the editors never set, so every campaign was saved
 * as email regardless of the channel the operator picked. It also invented a
 * description and a subject from the name, and stamped scheduledAt with the
 * current time on every save, which asserts that a draft is scheduled for now.
 *
 * Inventing those was a response to the contract marking them @required.
 * voxie#73 relaxed them to match what onion actually validates (name, channel,
 * content), so the client no longer has to fabricate anything.
 */
describe('campaign mutation payload (rtylr#938)', () => {
  const sent = (): Record<string, unknown> =>
    (mockClients.compatibility.createCampaign as ReturnType<typeof vi.fn>).mock
      .calls[0][0] as Record<string, unknown>;

  it('sends the channel the editor chose, not always email', async () => {
    vi.clearAllMocks();
    // SMSCampaign.tsx sends `type`, never `channel`.
    await crmApi.createCampaign({ name: 'Flash sale', type: 'sms', body: 'Today only' } as never);
    expect(sent().channel).toBe('sms');
  });

  it('does not invent a description from the subject or the name', async () => {
    vi.clearAllMocks();
    await crmApi.createCampaign({ name: 'Flash sale', type: 'sms', body: 'Today only' } as never);
    const description = sent().description;
    expect(description == null || description === '').toBe(true);
  });

  it('does not invent a subject for a channel that has none', async () => {
    vi.clearAllMocks();
    await crmApi.createCampaign({ name: 'Flash sale', type: 'sms', body: 'Today only' } as never);
    const subject = sent().subject;
    expect(subject == null || subject === '').toBe(true);
  });

  it('omits scheduledAt when nothing was scheduled', async () => {
    vi.clearAllMocks();
    await crmApi.createCampaign({ name: 'Flash sale', type: 'sms', body: 'Today only' } as never);
    const scheduledAt = sent().scheduled_at;
    expect(scheduledAt == null || scheduledAt === '').toBe(true);
  });

  it('passes a real description and subject through untouched', async () => {
    vi.clearAllMocks();
    await crmApi.createCampaign({
      name: 'Newsletter',
      type: 'email',
      subject: 'March news',
      description: 'Monthly roundup',
      body: 'Hello',
    } as never);
    expect(sent().description).toBe('Monthly roundup');
    expect(sent().subject).toBe('March news');
    expect(sent().channel).toBe('email');
    expect(sent().content).toBe('Hello');
  });

  it('passes a real schedule through', async () => {
    vi.clearAllMocks();
    await crmApi.createCampaign({
      name: 'Later',
      type: 'email',
      body: 'x',
      scheduledAt: '2026-09-01T10:00:00Z',
    } as never);
    expect(sent().scheduled_at).toBe('2026-09-01T10:00:00Z');
  });
});

/**
 * Automation mutation payload. The automations hook converts the UI record to
 * the canonical backend shape (triggerType/triggerConfig/actions/isActive)
 * before calling these methods (B53-R02). While this mapper only read the
 * legacy trigger/nodes/status keys, every create/update PUT collapsed to
 * {name, trigger_config:{}, is_active:false}: JSON.stringify dropped the
 * undefined trigger_type and actions, and onion's automations handler rejected
 * the body with 400 "Trigger type is required".
 */
describe('automation mutation payload (B53-R02)', () => {
  const canonical = {
    name: 'Welcome series',
    triggerType: 'customer_signup',
    triggerConfig: { sendDelayMinutes: 30 },
    actions: [{ type: 'send_email', template: 'welcome' }],
    isActive: true,
  };

  it('sends trigger_type and actions from the canonical payload on create', async () => {
    vi.clearAllMocks();
    await crmApi.createAutomation(canonical);
    expect(mockClients.compatibility.createAutomation).toHaveBeenCalledWith({
      name: 'Welcome series',
      trigger_type: 'customer_signup',
      trigger_config: { send_delay_minutes: 30 },
      actions: [{ type: 'send_email', template: 'welcome' }],
      is_active: true,
    });
  });

  it('sends trigger_type and actions from the canonical payload on update', async () => {
    vi.clearAllMocks();
    await crmApi.updateAutomation(9, canonical);
    expect(mockClients.compatibility.updateAutomation).toHaveBeenCalledWith({
      id: '9',
      body: {
        name: 'Welcome series',
        trigger_type: 'customer_signup',
        trigger_config: { send_delay_minutes: 30 },
        actions: [{ type: 'send_email', template: 'welcome' }],
        is_active: true,
      },
    });
  });

  it('still maps a legacy UI record', async () => {
    vi.clearAllMocks();
    await crmApi.updateAutomation(9, {
      status: 'paused',
      trigger: 'Birthday',
      nodes: [{ id: 'n2', type: 'action', label: 'Send email', x: 0, y: 80 }],
    });
    const call = (mockClients.compatibility.updateAutomation as ReturnType<typeof vi.fn>).mock
      .calls[0][0] as { body: Record<string, unknown> };
    expect(call.body.trigger_type).toBe('Birthday');
    expect(call.body.actions).toEqual([
      { id: 'n2', type: 'action', label: 'Send email', x: 0, y: 80 },
    ]);
    expect(call.body.is_active).toBe(false);
  });
});
