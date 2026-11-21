import { describe, expect, it } from 'vitest';
import { createSandboxPager } from './sandboxPager';

function sandbox() {
  let tick = Date.parse('2026-11-20T10:00:00.000Z');
  return createSandboxPager({ venueName: () => 'Demo Café', now: () => new Date((tick += 1000)) });
}

describe('sandbox pager', () => {
  it('links the phone and the board to one ticket, whichever side enters the number first', async () => {
    const pager = sandbox();
    const joined = await pager.joinTicket({ slug: 'demo', number: '٤٢' });
    expect(joined).toMatchObject({ number: '42', status: 'waiting' });

    const fromBoard = await pager.createTicket({ number: '42' });
    expect(fromBoard.ticket).toMatchObject({ number: '42', phoneLinked: true, source: 'customer' });
    expect(fromBoard.ticket).not.toHaveProperty('publicToken');

    const again = await pager.joinTicket({ slug: 'demo', number: '42' });
    expect(again.publicToken).toBe(joined.publicToken);
    expect((await pager.listTickets({})).tickets).toHaveLength(1);
  });

  it('rings the phone on ready and on every call again, until the customer answers', async () => {
    const pager = sandbox();
    const { publicToken } = await pager.joinTicket({ slug: 'demo', number: '7' });
    const id = (await pager.listTickets({})).tickets?.[0]?.id;

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

  it('counts the waiting orders entered before this one', async () => {
    const pager = sandbox();
    await pager.createTicket({ number: '1' });
    await pager.createTicket({ number: '2' });
    const { publicToken } = await pager.joinTicket({ slug: 'demo', number: '3' });
    expect((await pager.getPublicTicket({ publicToken })).ahead).toBe(2);

    const first = (await pager.listTickets({})).tickets?.[0]?.id;
    await pager.markReady({ id: first });
    expect((await pager.getPublicTicket({ publicToken })).ahead).toBe(1);
  });

  it('answers with the API error codes the screens act on', async () => {
    const pager = sandbox();
    await expect(pager.joinTicket({ slug: 'demo', number: '--' })).rejects.toMatchObject({
      statusCode: 400,
      code: 'invalid_number',
    });

    const { ticket } = await pager.createTicket({ number: '5' });
    await expect(pager.remindTicket({ id: ticket?.id })).rejects.toMatchObject({
      statusCode: 409,
      code: 'invalid_transition',
    });

    await pager.updateSettings({ strict: true });
    await expect(pager.joinTicket({ slug: 'demo', number: '6' })).rejects.toMatchObject({
      statusCode: 404,
      code: 'ticket_not_found',
    });

    await pager.updateSettings({ enabled: false });
    await expect(pager.getPublicVenue({ slug: 'demo' })).rejects.toMatchObject({
      statusCode: 404,
      code: 'venue_not_found',
    });
  });
});
