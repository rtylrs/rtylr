import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import {
  isValidServerUrl,
  isHostedPosOrigin,
  normalizeBaseUrl,
  resolveApiBaseUrl,
  SERVER_STORE_KEY,
  type ServerMode,
} from '../services/server/serverConfig';
import { fetchRuntimeConfig } from '../services/server/runtimeConfig';

interface ServerState {
  /** Which backend the app talks to. */
  mode: ServerMode;
  /** Self-host API base URL (normalized, trailing slash). Null when hosted. */
  customApiUrl: string | null;
  /** Auth (sign-in) host, from the self-host backend's runtime-config. */
  authUrl: string | null;
  /** Upload endpoint, from the self-host backend's runtime-config. */
  uploadUrl: string | null;
  /**
   * Whether the user has made an explicit choice. Drives the first-run
   * "Hosted vs Custom server" chooser — false means it hasn't been answered.
   */
  configured: boolean;
}

interface ServerActions {
  /** Use the hosted rtylr backend (the default). */
  useHosted: () => void;
  /**
   * Point the app at a customer's self-hosted backend. Validates the URL,
   * persists it, then best-effort fetches /runtime-config to learn the
   * customer's auth/upload URLs. Returns false if the URL is invalid.
   */
  useSelfHost: (apiUrl: string) => Promise<boolean>;
  /** Reset to the unconfigured/hosted defaults. */
  reset: () => void;
}

const initialState: ServerState = {
  mode: 'hosted',
  customApiUrl: null,
  authUrl: null,
  uploadUrl: null,
  configured: isHostedPosOrigin(),
};

const hostedOriginState = (state: ServerState & ServerActions): ServerState & ServerActions => ({
  ...state,
  mode: 'hosted',
  customApiUrl: null,
  authUrl: null,
  uploadUrl: null,
  configured: true,
});

export const useServerStore = create<ServerState & ServerActions>()(
  persist(
    (set) => ({
      ...initialState,

      useHosted: (): void => {
        set({
          mode: 'hosted',
          customApiUrl: null,
          authUrl: null,
          uploadUrl: null,
          configured: true,
        });
      },

      useSelfHost: async (apiUrl: string): Promise<boolean> => {
        if (!isValidServerUrl(apiUrl)) return false;
        const normalized = normalizeBaseUrl(apiUrl);
        // Persist the base immediately so resolvers pick it up even if the
        // runtime-config fetch is slow or fails.
        set({
          mode: 'self_host',
          customApiUrl: normalized,
          authUrl: null,
          uploadUrl: null,
          configured: true,
        });
        const { authUrl, uploadUrl } = await fetchRuntimeConfig(normalized);
        set({ authUrl, uploadUrl });
        return true;
      },

      reset: (): void => {
        set({ ...initialState });
      },
    }),
    {
      name: SERVER_STORE_KEY,
      merge: (persistedState, currentState) => {
        if (isHostedPosOrigin()) return hostedOriginState(currentState);
        return {
          ...currentState,
          ...(persistedState as Partial<ServerState> | null),
        };
      },
      partialize: (state) => ({
        mode: state.mode,
        customApiUrl: state.customApiUrl,
        authUrl: state.authUrl,
        uploadUrl: state.uploadUrl,
        configured: state.configured,
      }),
    }
  )
);

/** Re-exported so callers can read the active base without the store. */
export { resolveApiBaseUrl };
