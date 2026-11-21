import { describe, expect, it } from 'vitest';
import { createSandboxPager } from './sandboxPager';

function sandbox() {
  let tick = Date.parse('2026-11-20T10:00:00.000Z');
  return createSandboxPager({ venueName: () => 'Demo Café', now: () => new Date((tick += 1000)) });
}

async function scanAndLabel(pager: ReturnType<typeof sandbox>, label: string) {
  const { publicToken } = await pager.scanTicket({ slug: 'demo' });
  const tickets = (await pager.listTickets({})).tickets ?? [];
  const id = tickets[tickets.length - 1]?.id;
  await pager.labelTicket({ id, label });
  return { publicToken, id };
}

describe('sandbox pager', () => {
  it('gives each scanned phone its own code until the board labels it', async () => {
    const pager = sandbox();
    const codes = new Set<string>();
    for (let i = 0; i < 30; i += 1) {
      const scanned = await pager.scanTicket({ slug: 'demo' });
      expect(scanned.code).toMatch(/^[2-9A-HJ-NP-Z]{2}$/);
      codes.add(scanned.code ?? '');
    }
    expect(codes.size).toBe(30);

    const { publicToken } = await pager.scanTicket({ slug: 'demo' });
    const phone = await pager.getPublicTicket({ publicToken });
    expect(phone).toMatchObject({ status: 'waiting', ahead: 0 });
    expect(phone.label).toBeUndefined();

    const id = (await pager.listTickets({})).tickets?.at(-1)?.id;
    const { ticket } = await pager.labelTicket({ id, label: ' ٤٢ ' });
    expect(ticket).toMatchObject({ label: '42', labelKind: 'number' });
    expect(ticket).not.toHaveProperty('publicToken');
    expect(ticket?.code).toBeUndefined();
    expect(await pager.getPublicTicket({ publicToken })).toMatchObject({
      label: '42',
      labelKind: 'number',
    });
  });

  it('follows the label mode, and keeps an open number on one ticket', async () => {
    const pager = sandbox();
    await scanAndLabel(pager, '17');
    const { publicToken } = await pager.scanTicket({ slug: 'demo' });
    const id = (await pager.listTickets({})).tickets?.at(-1)?.id;

    await expect(pager.labelTicket({ id, label: '17' })).rejects.toMatchObject({
      statusCode: 409,
      code: 'label_taken',
    });
    expect((await pager.labelTicket({ id, label: 'Sara  Haddad' })).ticket).toMatchObject({
      label: 'Sara Haddad',
      labelKind: 'name',
    });

    await pager.updateSettings({ labelMode: 'number' });
    await expect(pager.labelTicket({ id, label: 'Sara' })).rejects.toMatchObject({
      statusCode: 400,
      code: 'invalid_label',
    });
    await pager.updateSettings({ labelMode: 'name' });
    expect((await pager.labelTicket({ id, label: '17' })).ticket).toMatchObject({
      labelKind: 'name',
    });
    expect((await pager.getPublicTicket({ publicToken })).label).toBe('17');
  });

  it('rings the phone on ready and on every call again, until the customer answers', async () => {
    const pager = sandbox();
    const scanned = await pager.scanTicket({ slug: 'demo' });
    const id = (await pager.listTickets({})).tickets?.[0]?.id;
    const { publicToken } = scanned;

    await expect(pager.markReady({ id })).rejects.toMatchObject({
      statusCode: 409,
      code: 'invalid_transition',
    });
    await pager.labelTicket({ id, label: '7' });
    await pager.markReady({ id });
    const ready = await pager.getPublicTicket({ publicToken });
    expect(ready).toMatchObject({ status: 'ready', venueName: 'Demo Café' });
    expect(ready.readyAt).toBeDefined();

    await pager.acknowledgeTicket({ publicToken });
    expect((await pager.getPublicTicket({ publicToken })).acknowledgedAt).toBeDefined();

    await pager.remindTicket({ id });
    const reminded = await pager.getPublicTicket({ publicToken });
    expect(reminded.acknowledgedAt).toBeUndefined();
    expect(reminded.remindedAt).toBeDefined();

    await pager.collectTicket({ id });
    expect((await pager.getPublicTicket({ publicToken })).status).toBe('collected');
    expect((await pager.listTickets({})).tickets).toEqual([]);
  });

  it('counts the labelled waiting orders that scanned before this one', async () => {
    const pager = sandbox();
    await scanAndLabel(pager, '1');
    await pager.scanTicket({ slug: 'demo' });
    await scanAndLabel(pager, '2');
    const { publicToken } = await scanAndLabel(pager, '3');
    expect((await pager.getPublicTicket({ publicToken })).ahead).toBe(2);

    const first = (await pager.listTickets({})).tickets?.[0]?.id;
    await pager.markReady({ id: first });
    expect((await pager.getPublicTicket({ publicToken })).ahead).toBe(1);
  });

  it('answers not found once the demo venue switches its pager off', async () => {
    const pager = sandbox();
    await pager.updateSettings({ enabled: false });
    await expect(pager.scanTicket({ slug: 'demo' })).rejects.toMatchObject({
      statusCode: 404,
      code: 'venue_not_found',
    });
  });
});
