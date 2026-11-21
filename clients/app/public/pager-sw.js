// Service worker for the order pager, scoped to the ticket pages: shows the
// "your order is ready" push on a locked phone. It has no fetch handler, so it
// never serves or caches pages. The ticket page stores the notification text
// in the visitor's language before it subscribes; see
// src/modules/pager/shared/lib/pushAlerts.ts.

const PUSH_TEXT_CACHE = 'pager-push-text';
const VIBRATION = [600, 250, 600, 250, 600, 250, 600];

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (event) => event.waitUntil(self.clients.claim()));

function ticketPath(token) {
  return `/pager/t/${token}`;
}

async function textFor(message) {
  try {
    const cache = await caches.open(PUSH_TEXT_CACHE);
    const stored = await cache.match(ticketPath(message.token));
    const text = stored ? await stored.json() : null;
    const kind = text && text[message.kind];
    if (kind && kind.title) return kind;
  } catch {
    // Fall through to the plain text the push itself carries.
  }
  return { title: message.number, body: message.venue || '' };
}

self.addEventListener('push', (event) => {
  let message = null;
  try {
    message = event.data ? event.data.json() : null;
  } catch {
    message = null;
  }
  if (!message || typeof message.token !== 'string') return;
  event.waitUntil(
    textFor(message).then((text) =>
      self.registration.showNotification(text.title, {
        body: text.body,
        // One notification per ticket: a call again replaces the ready one
        // and, with renotify, rings and vibrates again.
        tag: `pager-${message.token}`,
        renotify: true,
        requireInteraction: true,
        vibrate: VIBRATION,
        icon: '/favicon/android-chrome-192x192.png',
        badge: '/favicon/android-chrome-192x192.png',
        data: { url: ticketPath(message.token) },
      })
    )
  );
});

self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const url = new URL(event.notification.data?.url || '/', self.location.origin).href;
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windows) => {
      const open = windows.find((client) => client.url === url);
      if (open) return open.focus();
      return self.clients.openWindow(url);
    })
  );
});
