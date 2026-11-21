import { vi } from 'vitest';

export interface FakePushBrowser {
  requestPermission: ReturnType<typeof vi.fn>;
  register: ReturnType<typeof vi.fn>;
  subscribe: ReturnType<typeof vi.fn>;
  unsubscribe: ReturnType<typeof vi.fn>;
  setPermission: (permission: NotificationPermission) => void;
  /** A subscription this browser already holds, made with applicationServerKey. */
  setExisting: (applicationServerKey: Uint8Array | null) => void;
  /** What the page left in the service worker's text cache, by request URL. */
  storedText: (url: string) => Promise<unknown>;
  restore: () => void;
}

export const FAKE_SUBSCRIPTION = {
  endpoint: 'https://fcm.googleapis.com/fcm/send/fake',
  keys: { p256dh: 'BPfake', auth: 'authfake' },
};

/**
 * A browser with Web Push: Notification, PushManager and a service worker
 * container whose registration is already active.
 */
export function installFakePushBrowser(): FakePushBrowser {
  let permission: NotificationPermission = 'default';
  let existing: { options: { applicationServerKey: ArrayBuffer | null } } | null = null;
  const unsubscribe = vi.fn(async () => {
    existing = null;
    return true;
  });
  const subscribe = vi.fn(async (options: { applicationServerKey: Uint8Array }) => ({
    options: { applicationServerKey: options.applicationServerKey.buffer },
    toJSON: () => FAKE_SUBSCRIPTION,
    unsubscribe,
  }));
  const registration = {
    active: {},
    pushManager: {
      subscribe,
      getSubscription: vi.fn(async () =>
        existing ? { ...existing, toJSON: () => FAKE_SUBSCRIPTION, unsubscribe } : null
      ),
    },
  };
  const register = vi.fn(async () => registration);
  const requestPermission = vi.fn(async () => permission);
  const cacheEntries = new Map<string, Response>();
  const cache = {
    put: vi.fn(async (url: string, response: Response) => {
      cacheEntries.set(url, response);
    }),
    keys: vi.fn(async () => [...cacheEntries.keys()]),
    delete: vi.fn(async (url: string) => cacheEntries.delete(url)),
  };

  vi.stubGlobal('Notification', {
    get permission() {
      return permission;
    },
    requestPermission,
  });
  vi.stubGlobal('PushManager', function PushManager() {});
  vi.stubGlobal('caches', { open: vi.fn(async () => cache) });
  Object.defineProperty(navigator, 'serviceWorker', {
    configurable: true,
    value: { register },
  });

  return {
    requestPermission,
    register,
    subscribe,
    unsubscribe,
    setPermission: (next) => {
      permission = next;
    },
    setExisting: (key) => {
      existing = key ? { options: { applicationServerKey: key.slice().buffer } } : null;
    },
    storedText: async (url) => cacheEntries.get(url)?.clone().json(),
    restore: () => {
      vi.unstubAllGlobals();
      delete (navigator as { serviceWorker?: unknown }).serviceWorker;
    },
  };
}
