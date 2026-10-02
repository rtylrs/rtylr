import { createJSONStorage } from 'zustand/middleware';

/**
 * Must match the persist key of `useUserStore` (apps/portal/src/shared/stores/userStore.ts).
 * The persisted payload is zustand's `{ state: { user }, version }` envelope, and
 * `user.businessId` scopes dashboard layout/menu storage per business.
 */
const USER_STORE_PERSIST_KEY = 'user-store';

export function getTenantKey(baseKey: string): string {
  try {
    const rawUser = localStorage.getItem(USER_STORE_PERSIST_KEY);
    if (rawUser) {
      const parsed = JSON.parse(rawUser);
      const bId = parsed?.state?.user?.businessId;
      if (bId) return `${baseKey}:${bId}`;
    }
  } catch {
    // ignore parse errors
  }
  return baseKey;
}

export const tenantStorage = createJSONStorage(() => ({
  getItem: (name: string) => localStorage.getItem(getTenantKey(name)),
  setItem: (name: string, value: string) => localStorage.setItem(getTenantKey(name), value),
  removeItem: (name: string) => localStorage.removeItem(getTenantKey(name)),
}));
