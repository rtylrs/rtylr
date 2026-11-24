import { useEffect, useState } from 'react';
import { PAGER_TICKET_PREFIX } from '../../pagerPaths';
import { publicPagerClient } from './clients';
import { usePagerSandbox } from './pagerClients';

// Lock-screen notifications for a ticket, on browsers with Web Push (Android
// Chrome and desktop browsers; an iPhone Safari tab has none). They come on
// top of the in-page bell, so every failure here is quiet: the page still rings.

export const PAGER_SW_URL = '/pager-sw.js';
/** Only the ticket pages: the worker never sees the staff board or the Portal. */
export const PAGER_SW_SCOPE = PAGER_TICKET_PREFIX;
/** Read by entrypoints/rennili/public/pager-sw.js; keep the two in step. */
export const PUSH_TEXT_CACHE = 'pager-push-text';
const MAX_STORED_TEXTS = 20;
const NOTIFICATION_VIBRATION = [600, 250, 600, 250, 600, 250, 600];

export interface PushMessageText {
  title: string;
  body: string;
}

/** The notification text in the visitor's language, one per kind of push. */
export interface PushText {
  ready: PushMessageText;
  remind: PushMessageText;
}

export function pushSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    'serviceWorker' in navigator &&
    'PushManager' in window &&
    'Notification' in window
  );
}

export function notificationPermission(): NotificationPermission | 'unsupported' {
  if (!pushSupported()) return 'unsupported';
  return Notification.permission;
}

/**
 * An iPhone or iPad browser tab. iOS gives Web Push only to a site added to the
 * home screen, so a tab rings only while it is on screen. iPadOS reports a Mac,
 * told apart by its touch screen.
 */
export function iosTabWithoutPush(): boolean {
  if (typeof navigator === 'undefined' || pushSupported()) return false;
  const agent = navigator.userAgent;
  return (
    /iPhone|iPad|iPod/.test(agent) || (/Macintosh/.test(agent) && navigator.maxTouchPoints > 1)
  );
}

/**
 * An Android in-app QR scanner or social WebView rather than full Chrome.
 * These embedded browsers commonly block both `navigator.vibrate()` and Web Push.
 */
export function androidWebView(): boolean {
  if (typeof navigator === 'undefined') return false;
  const agent = navigator.userAgent;
  if (!/Android/i.test(agent)) return false;
  if (/;\s*wv\)|FBAN|FBAV|Instagram|Line\/|MicroMessenger|Snapchat/i.test(agent)) return true;
  return !pushSupported();
}

/** Android intent URL that hands the current HTTPS page directly to Chrome. */
export function openInChromeUrl(href?: string): string | null {
  const target = href ?? (typeof window !== 'undefined' ? window.location.href : '');
  if (!target) return null;
  try {
    const parsed = new URL(target);
    if (parsed.protocol !== 'https:' && parsed.protocol !== 'http:') return null;
    const scheme = parsed.protocol.slice(0, -1);
    return `intent://${parsed.host}${parsed.pathname}${parsed.search}${parsed.hash}#Intent;scheme=${scheme};package=com.android.chrome;end`;
  } catch {
    return null;
  }
}

let pushKey: Promise<string> | null = null;

/** The server's VAPID key, fetched once per document; empty while push is off. */
export function loadPushKey(): Promise<string> {
  if (!pushSupported()) return Promise.resolve('');
  pushKey ??= publicPagerClient.getPushKey({}).then(
    (response) => response.publicKey ?? '',
    () => {
      pushKey = null;
      return '';
    }
  );
  return pushKey;
}

/**
 * The key for the screens. Empty in the public demo: its tickets live only in
 * the browser, so the server has nothing to push to and the visitor is never
 * asked. An empty key keeps askPushPermission and followTicketByPush silent.
 */
export function usePushKey(): string {
  const sandbox = usePagerSandbox();
  const [key, setKey] = useState('');
  useEffect(() => {
    if (sandbox) return undefined;
    let current = true;
    void loadPushKey().then((loaded) => {
      if (current) setKey(loaded);
    });
    return () => {
      current = false;
    };
  }, [sandbox]);
  return sandbox ? '' : key;
}

let pendingPermission: Promise<NotificationPermission> | null = null;

/**
 * Requests browser notification permission synchronously inside a tap handler,
 * without waiting for the VAPID public key network round-trip. Returns the
 * in-flight permission promise so callers can refresh UI once the user answers.
 */
export function askPushPermissionNow(): Promise<NotificationPermission> | null {
  if (!pushSupported()) return null;
  if (pendingPermission) return pendingPermission;
  if (Notification.permission !== 'default') return Promise.resolve(Notification.permission);
  pendingPermission = Notification.requestPermission()
    .catch((): NotificationPermission => 'default')
    .finally(() => {
      pendingPermission = null;
    });
  return pendingPermission;
}

/**
 * Asks for notification permission. Browsers only show the prompt from a tap,
 * so this runs synchronously in the click handler, before any await. It asks
 * only when the server has push on and the visitor has not answered yet.
 */
export function askPushPermission(publicKey: string): void {
  if (!publicKey) return;
  void askPushPermissionNow();
}

function permissionSettled(): Promise<NotificationPermission> {
  return pendingPermission ?? Promise.resolve(Notification.permission);
}

function keyBytes(base64Url: string): Uint8Array<ArrayBuffer> {
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64.padEnd(Math.ceil(base64.length / 4) * 4, '='));
  const bytes = new Uint8Array(new ArrayBuffer(binary.length));
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

function sameKey(current: ArrayBuffer | null, wanted: Uint8Array): boolean {
  if (!current) return false;
  const bytes = new Uint8Array(current);
  return bytes.length === wanted.length && bytes.every((byte, i) => byte === wanted[i]);
}

function activated(registration: ServiceWorkerRegistration): Promise<void> {
  const worker = registration.installing ?? registration.waiting;
  if (registration.active || !worker) return Promise.resolve();
  return new Promise((resolve) => {
    // A worker that failed to install turns redundant; subscribing then fails.
    worker.addEventListener('statechange', () => {
      if (worker.state === 'activated' || worker.state === 'redundant') resolve();
    });
  });
}

/** Leaves the text where the service worker looks for it, keyed by ticket. */
async function storeText(token: string, text: PushText): Promise<void> {
  if (typeof caches === 'undefined') return;
  const cache = await caches.open(PUSH_TEXT_CACHE);
  await cache.put(
    `${PAGER_SW_SCOPE}${token}`,
    new Response(JSON.stringify(text), { headers: { 'Content-Type': 'application/json' } })
  );
  const stored = await cache.keys();
  await Promise.all(
    stored.slice(0, Math.max(0, stored.length - MAX_STORED_TEXTS)).map((old) => cache.delete(old))
  );
}

/**
 * Subscribes this browser to the ticket's pushes once the visitor has granted
 * permission, waiting for a prompt that is still open. Safe to call again: the
 * server keeps one subscription per browser and ticket.
 */
export async function followTicketByPush(options: {
  publicKey: string;
  token: string;
  text: PushText;
}): Promise<boolean> {
  const { publicKey, token, text } = options;
  if (!publicKey || !token || !pushSupported()) return false;
  if ((await permissionSettled()) !== 'granted') return false;

  const registration = await navigator.serviceWorker.register(PAGER_SW_URL, {
    scope: PAGER_SW_SCOPE,
  });
  await activated(registration);
  await storeText(token, text);

  const applicationServerKey = keyBytes(publicKey);
  let subscription = await registration.pushManager.getSubscription();
  if (subscription && !sameKey(subscription.options.applicationServerKey, applicationServerKey)) {
    // The server's key changed since this browser subscribed.
    await subscription.unsubscribe();
    subscription = null;
  }
  subscription ??= await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey,
  });

  const { endpoint, keys } = subscription.toJSON();
  if (!endpoint || !keys?.p256dh || !keys.auth) return false;
  await publicPagerClient.subscribePush({
    publicToken: token,
    endpoint,
    p256dh: keys.p256dh,
    auth: keys.auth,
  });
  return true;
}

/**
 * Triggers a local Service Worker notification for an order that just became
 * ready. On Android Chrome this routes vibration through the OS notification
 * channel (working even when Touch Feedback haptics are disabled or the tab
 * was reloaded), and shares the `pager-<token>` tag with server Web Push so
 * notifications never duplicate.
 */
export async function showTicketNotification(options: {
  token: string;
  title: string;
  body: string;
}): Promise<boolean> {
  const { token, title, body } = options;
  if (!token || !title || !pushSupported()) return false;
  if (Notification.permission !== 'granted') return false;
  try {
    const registration = await navigator.serviceWorker.register(PAGER_SW_URL, {
      scope: PAGER_SW_SCOPE,
    });
    await activated(registration);
    if (typeof registration.showNotification !== 'function') return false;
    await registration.showNotification(title, {
      body,
      tag: `pager-${token}`,
      renotify: true,
      requireInteraction: true,
      vibrate: NOTIFICATION_VIBRATION,
      icon: '/favicon/android-chrome-192x192.png',
      badge: '/favicon/android-chrome-192x192.png',
      data: { url: `${PAGER_SW_SCOPE}${token}` },
    } as NotificationOptions);
    return true;
  } catch {
    return false;
  }
}

/** Dismisses any active notification for this ticket once collected or acknowledged. */
export async function closeTicketNotification(token: string): Promise<void> {
  if (!token || !pushSupported()) return;
  try {
    const sw = navigator.serviceWorker;
    if (typeof sw.getRegistration !== 'function') return;
    const registration = await sw.getRegistration(PAGER_SW_SCOPE);
    if (!registration || typeof registration.getNotifications !== 'function') return;
    const active = await registration.getNotifications({ tag: `pager-${token}` });
    for (const notification of active) notification.close();
  } catch {
    // Best-effort cleanup.
  }
}
