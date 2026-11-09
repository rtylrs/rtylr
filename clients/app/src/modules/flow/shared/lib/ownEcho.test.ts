import { describe, it, expect } from 'vitest';
import { MutationObserver, QueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

import { isOwnEcho } from './ownEcho';

const mine: RealtimeEvent = { topic: 'project:1', type: 'task.updated', actor_id: 42 };

async function changeHere(client: QueryClient): Promise<void> {
  await new MutationObserver(client, { mutationFn: async () => undefined }).mutate();
}

describe('isOwnEcho', () => {
  // The same person's other tab or device made no change of its own, so the
  // event is news there and has to be applied.
  it('applies my event in a tab that changed nothing', () => {
    expect(isOwnEcho(new QueryClient(), mine, '42')).toBe(false);
  });

  it('skips my event in the tab that just made the change', async () => {
    const client = new QueryClient();
    await changeHere(client);
    expect(isOwnEcho(client, mine, '42')).toBe(true);
  });

  it('skips my event while this tab still has a change in flight', () => {
    const client = new QueryClient();
    void new MutationObserver(client, { mutationFn: () => new Promise(() => {}) }).mutate();
    expect(isOwnEcho(client, mine, '42')).toBe(true);
  });

  it('applies my event once this tab’s last change is old', async () => {
    const client = new QueryClient();
    await changeHere(client);
    expect(isOwnEcho(client, mine, '42', Date.now() + 10_000)).toBe(false);
  });

  it('applies a teammate’s event even right after a change here', async () => {
    const client = new QueryClient();
    await changeHere(client);
    expect(isOwnEcho(client, { ...mine, actor_id: 999 }, '42')).toBe(false);
  });

  it('applies events with no actor or no known user', async () => {
    const client = new QueryClient();
    await changeHere(client);
    expect(isOwnEcho(client, { ...mine, actor_id: undefined }, '42')).toBe(false);
    expect(isOwnEcho(client, mine, null)).toBe(false);
  });
});
