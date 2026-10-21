import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { createOneKitClient } from '@voxire/api';
import { parseEffectiveAccess } from '@voxire/auth';
import { OauthServiceClient } from '@voxire/voxie/common/oauth/v1/client';
import { ConfigServiceClient } from '@voxire/voxie/hub/config/v1/client';
import { SettingsServiceClient } from '@voxire/voxie/pos/settings/v1/client';

type EffectiveAccess = NonNullable<ReturnType<typeof parseEffectiveAccess>>;

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

interface ApiError extends Error {
  response?: { data?: { message?: string } };
}

const oauthClient = createOneKitClient(OauthServiceClient, 'pos');
const configClient = createOneKitClient(ConfigServiceClient, 'pos');
const settingsClient = createOneKitClient(SettingsServiceClient, 'pos');

export interface StoredUser {
  effectiveAccess?: EffectiveAccess | null;
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
  effectiveAccess: EffectiveAccess | null;
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
  effectiveAccess: null,
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

export function businessContactPhone(
  profile: BusinessProfile | null | undefined
): string | undefined {
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
        set({
          user: null,
          isAuthenticated: false,
          loading: false,
          error: null,
          effectiveAccess: null,
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
          const [me, config] = await Promise.all([
            oauthClient.me({}),
            configClient.getConfig({}).catch(() => ({ business: undefined })),
          ]);
          const state = get();
          const profile = normalizeBusinessProfile(config.business);
          const effectiveAccess = parseEffectiveAccess(me.effectiveAccess);
          const user: StoredUser = {
            ...(state.user ?? {}),
            id: me.userId,
            effectiveAccess,
          };
          set({
            effectiveAccess,
            userId: idText(me.userId) ?? state.userId,
            token: state.token,
            user,
            businessProfile: profile ?? state.businessProfile,
            isAuthenticated: true,
            loading: false,
            addons: user.addons ?? state.addons,
            visits: state.visits,
          });
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || error.message || 'Failed to restore session',
          });
          throw err;
        }
      },

      setThemeAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const current = await settingsClient.getPOSSettings({});
          const response = await settingsClient.updatePOSSettings({
            businessProfile: current.settings?.businessProfile ?? '',
            theme: JSON.stringify(payload.themeData),
            features: current.settings?.features,
            receiptFooter: current.settings?.receiptFooter ?? '',
          });
          const state = get();
          if (state.user) {
            set({
              user: {
                ...state.user,
                theme: response.settings?.theme
                  ? JSON.parse(response.settings.theme)
                  : payload.themeData,
              },
              loading: false,
              error: null,
            });
          }
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || error.message || 'Failed to update theme',
          });
          throw err;
        }
      },
    }),
    {
      name: 'pos-user-store',
      version: 1,
      partialize: (state) => ({
        user: state.user,
        businessProfile: state.businessProfile,
        // The JWT is deliberately NOT persisted here - the httpOnly-style
        // token cookie is the single persistent credential, and duplicating
        // it into localStorage exposes it to any XSS.
        isAuthenticated: state.isAuthenticated,
        effectiveAccess: state.effectiveAccess,
        userId: state.userId,
        addons: state.addons,
        visits: state.visits,
      }),
    }
  )
);
