import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  FAKE_SUBSCRIPTION,
  installFakePushBrowser,
  type FakePushBrowser,
} from '../testing/fakePushBrowser';

const api = vi.hoisted(() => ({ getPushKey: vi.fn(), subscribePush: vi.fn() }));
vi.mock('./clients', () => ({ publicPagerClient: api }));

// URL-safe base64 of the bytes 1..65, the shape of a VAPID public key.
const KEY_BYTES = Uint8Array.from({ length: 65 }, (_, i) => i + 1);
const PUBLIC_KEY = btoa(String.fromCharCode(...KEY_BYTES))
  .replace(/\+/g, '-')
  .replace(/\//g, '_')
  .replace(/=+$/, '');
const TEXT = {
  ready: { title: 'Order 42 is ready', body: 'Collect it at the counter of Acme.' },
  remind: { title: 'Order 42 is waiting for you', body: 'Acme is calling your number again.' },
};

let browser: FakePushBrowser;

// Each test gets fresh module state: the cached key and any open prompt.
async function load() {
  vi.resetModules();
  return import('./pushAlerts');
}

beforeEach(() => {
  browser = installFakePushBrowser();
  api.getPushKey.mockReset().mockResolvedValue({ publicKey: PUBLIC_KEY });
  api.subscribePush.mockReset().mockResolvedValue({});
});

afterEach(() => {
  browser.restore();
});

describe('asking for notification permission', () => {
  it('asks synchronously, inside the tap, when the server has push on', async () => {
    const { askPushPermission } = await load();

    askPushPermission(PUBLIC_KEY);

    expect(browser.requestPermission).toHaveBeenCalledTimes(1);
  });

  it('stays silent without a server key', async () => {
    const { askPushPermission } = await load();

    askPushPermission('');

    expect(browser.requestPermission).not.toHaveBeenCalled();
  });

  it('does not ask again once the visitor has answered', async () => {
    const { askPushPermission } = await load();
    for (const answered of ['granted', 'denied'] as const) {
      browser.setPermission(answered);
      askPushPermission(PUBLIC_KEY);
    }
    expect(browser.requestPermission).not.toHaveBeenCalled();
  });

  it('stays silent on a browser without Web Push, such as an iPhone Safari tab', async () => {
    browser.restore();
    const { askPushPermission, loadPushKey } = await load();

    askPushPermission(PUBLIC_KEY);

    expect(await loadPushKey()).toBe('');
    expect(api.getPushKey).not.toHaveBeenCalled();
  });
});

describe('following a ticket by push', () => {
  it('subscribes with the server key once permission is granted and tells the server', async () => {
    browser.setPermission('granted');
    const { followTicketByPush } = await load();

    await expect(
      followTicketByPush({ publicKey: PUBLIC_KEY, token: 'tok-1', text: TEXT })
    ).resolves.toBe(true);

    expect(browser.register).toHaveBeenCalledWith('/pager-sw.js', { scope: '/pager/t/' });
    expect(browser.subscribe).toHaveBeenCalledWith({
      userVisibleOnly: true,
      applicationServerKey: KEY_BYTES,
    });
    expect(api.subscribePush).toHaveBeenCalledWith({
      publicToken: 'tok-1',
      endpoint: FAKE_SUBSCRIPTION.endpoint,
      p256dh: FAKE_SUBSCRIPTION.keys.p256dh,
      auth: FAKE_SUBSCRIPTION.keys.auth,
    });
    // Where public/pager-sw.js reads the text in the visitor's language.
    expect(await browser.storedText('/pager/t/tok-1')).toEqual(TEXT);
  });

  it('waits for the prompt opened by the tap', async () => {
    const { askPushPermission, followTicketByPush } = await load();
    let answer: (permission: NotificationPermission) => void = () => undefined;
    browser.requestPermission.mockReturnValue(new Promise((resolve) => (answer = resolve)));

    askPushPermission(PUBLIC_KEY);
    const following = followTicketByPush({ publicKey: PUBLIC_KEY, token: 'tok-1', text: TEXT });
    await Promise.resolve();
    expect(browser.register).not.toHaveBeenCalled();

    browser.setPermission('granted');
    answer('granted');
    await expect(following).resolves.toBe(true);
    expect(api.subscribePush).toHaveBeenCalledTimes(1);
  });

  it('does nothing when permission is refused or was never asked', async () => {
    const { followTicketByPush } = await load();
    for (const permission of ['default', 'denied'] as const) {
      browser.setPermission(permission);
      await expect(
        followTicketByPush({ publicKey: PUBLIC_KEY, token: 'tok-1', text: TEXT })
      ).resolves.toBe(false);
    }
    expect(browser.register).not.toHaveBeenCalled();
    expect(api.subscribePush).not.toHaveBeenCalled();
  });

  it('does nothing while the server has push off', async () => {
    browser.setPermission('granted');
    const { followTicketByPush } = await load();

    await expect(followTicketByPush({ publicKey: '', token: 'tok-1', text: TEXT })).resolves.toBe(
      false
    );

    expect(browser.register).not.toHaveBeenCalled();
  });

  it('reuses this browser subscription, and replaces one made with an old server key', async () => {
    browser.setPermission('granted');
    const { followTicketByPush } = await load();

    browser.setExisting(KEY_BYTES);
    await followTicketByPush({ publicKey: PUBLIC_KEY, token: 'tok-1', text: TEXT });
    expect(browser.subscribe).not.toHaveBeenCalled();
    expect(api.subscribePush).toHaveBeenCalledTimes(1);

    browser.setExisting(new Uint8Array(65));
    await followTicketByPush({ publicKey: PUBLIC_KEY, token: 'tok-1', text: TEXT });
    expect(browser.unsubscribe).toHaveBeenCalledTimes(1);
    expect(browser.subscribe).toHaveBeenCalledTimes(1);
  });
});

describe('the screens push key', () => {
  async function renderKey(sandbox: boolean) {
    const { usePushKey } = await load();
    // After load(): the same pagerClients module the hook reads.
    const { PagerClientsProvider } = await import('./pagerClients');
    const value = { publicClient: api as never, staffClient: api as never, sandbox };
    return renderHook(() => usePushKey(), {
      wrapper: ({ children }: { children: React.ReactNode }) =>
        React.createElement(PagerClientsProvider, { value }, children),
    });
  }

  it('loads the server key outside the demo', async () => {
    const { result } = await renderKey(false);

    await waitFor(() => expect(result.current).toBe(PUBLIC_KEY));
  });

  it('stays empty in the demo, without asking the server', async () => {
    const { result } = await renderKey(true);

    await act(async () => {
      await Promise.resolve();
    });
    expect(result.current).toBe('');
    expect(api.getPushKey).not.toHaveBeenCalled();
  });
});
