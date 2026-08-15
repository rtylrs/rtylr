import { createJSONStorage } from 'zustand/middleware';

export function getTenantKey(baseKey: string): string {
  try {
    const rawUser = localStorage.getItem('user-storage');
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
