import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api/api';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

interface ApiError extends Error {
  response?: { data?: { message?: string } };
}

export interface StoredUser {
  role?: string;
  addons?: Addons;
  name?: string;
  email?: string;
  theme?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface BusinessProfile {
  id?: string;
  name?: string;
  siteName?: string;
  site_name?: string;
  logoUrl?: string;
  logo_url?: string;
  contactPhone?: string;
  contact_phone?: string;
  address?: string;
  currency?: string;
  timezone?: string;
  [key: string]: unknown;
}

interface UserState {
  user: StoredUser | null;
  businessProfile: BusinessProfile | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: unknown[];
}

interface UserActions {
  logout: () => void;
  clearError: () => void;
  setToken: (token: string) => void;
  setUser: (user: StoredUser) => void;
  getUserInfo: () => Promise<void>;
  setThemeAsync: (payload: { id: string; themeData: Record<string, unknown> }) => Promise<void>;
}

const initialState: UserState = {
  user: null,
  businessProfile: null,
  token: null,
  isAuthenticated: false,
  loading: false,
  error: null,
  role: null,
  userId: null,
  addons: {
    feedback: false,
    about: false,
    contact: false,
    landing: false,
  },
  visits: [],
};

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function text(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value.trim() : undefined;
}

function idText(value: unknown): string | null {
  if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
    return String(Math.floor(value));
  }
  return text(value) ?? null;
}

function normalizeBusinessProfile(raw: unknown): BusinessProfile | null {
  const source = record(raw);
  if (!source) return null;

  const name = text(source.name) ?? text(source.siteName) ?? text(source.site_name);
  if (!name) return null;

  return {
    ...source,
    id: idText(source.id) ?? undefined,
    name,
    siteName: text(source.siteName) ?? text(source.site_name) ?? name,
    site_name: text(source.site_name) ?? text(source.siteName) ?? name,
    logoUrl: text(source.logoUrl) ?? text(source.logo_url),
    logo_url: text(source.logo_url) ?? text(source.logoUrl),
    contactPhone: text(source.contactPhone) ?? text(source.contact_phone),
    contact_phone: text(source.contact_phone) ?? text(source.contactPhone),
    address: text(source.address),
    currency: text(source.currency),
    timezone: text(source.timezone),
  };
}

export function businessDisplayName(profile: BusinessProfile | null | undefined): string | null {
  return text(profile?.name) ?? text(profile?.siteName) ?? text(profile?.site_name) ?? null;
}

export function businessContactPhone(profile: BusinessProfile | null | undefined): string | undefined {
  return text(profile?.contactPhone) ?? text(profile?.contact_phone);
}

export function businessAddress(profile: BusinessProfile | null | undefined): string | undefined {
  return text(profile?.address);
}

export const useUserStore = create<UserState & UserActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      logout: () => {
        localStorage.removeItem('token');
        set({
          user: null,
          isAuthenticated: false,
          loading: false,
          error: null,
          role: null,
          userId: null,
          token: null,
        });
      },

      clearError: () => set({ error: null }),

      setToken: (token) => set({ token }),

      setUser: (user) => set({ user }),

      getUserInfo: async () => {
        set({ loading: true, error: null });
        try {
          const response = await api.get('/business/me');
          const data = record(response.data) ?? {};
          const state = get();
          const user = (record(data.user) ?? data) as StoredUser;
          const profile = normalizeBusinessProfile(data.business ?? data.profile ?? data);
          set({
            role: user.role ?? state.role,
            userId: idText(data.userId ?? data.user_id) ?? state.userId,
            token: text(data.token) ?? state.token,
            user,
            businessProfile: profile ?? state.businessProfile,
            isAuthenticated: true,
            loading: false,
            addons: user.addons ?? state.addons,
            visits: Array.isArray(data.visits) ? data.visits : state.visits,
          });
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || 'Failed to restore session',
          });
          throw err;
        }
      },

      setThemeAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const response = await api.put(`/business/${payload.id}/theme`, payload.themeData);
          const state = get();
          if (state.user) {
            set({
              user: { ...state.user, theme: response.data.theme },
              loading: false,
              error: null,
            });
          }
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || 'Failed to update theme',
          });
          throw err;
        }
      },
    }),
    {
      name: 'user-store',
      partialize: (state) => ({
        user: state.user,
        businessProfile: state.businessProfile,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        userId: state.userId,
        addons: state.addons,
        visits: state.visits,
      }),
    }
  )
);
