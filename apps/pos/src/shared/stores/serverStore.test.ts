import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest';
import { useServerStore } from './serverStore';
import { fetchRuntimeConfig } from '../services/server/runtimeConfig';

vi.mock('../services/server/runtimeConfig', () => ({
  fetchRuntimeConfig: vi.fn(),
}));

describe('serverStore', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useServerStore.getState().reset();
  });

  it('starts with initial unconfigured hosted state', () => {
    const state = useServerStore.getState();
    expect(state.mode).toBe('hosted');
    expect(state.customApiUrl).toBeNull();
    expect(state.authUrl).toBeNull();
    expect(state.uploadUrl).toBeNull();
    expect(state.configured).toBe(false); // since we are not on pos.rtylr.com in tests
  });

  describe('useHosted', () => {
    it('sets state to configured hosted mode', () => {
      useServerStore.getState().useHosted();
      const state = useServerStore.getState();
      expect(state.mode).toBe('hosted');
      expect(state.customApiUrl).toBeNull();
      expect(state.authUrl).toBeNull();
      expect(state.uploadUrl).toBeNull();
      expect(state.configured).toBe(true);
    });
  });

  describe('useSelfHost', () => {
    it('returns false and does not change state if URL is invalid', async () => {
      const success = await useServerStore.getState().useSelfHost('not-a-valid-url');
      expect(success).toBe(false);

      const state = useServerStore.getState();
      expect(state.mode).toBe('hosted');
      expect(state.configured).toBe(false);
    });

    it('sets self-host parameters and fetches runtime config on success', async () => {
      (fetchRuntimeConfig as Mock).mockResolvedValue({
        authUrl: 'https://auth.customer.com',
        uploadUrl: 'https://upload.customer.com',
      });

      const success = await useServerStore.getState().useSelfHost('http://192.168.1.50:8000');
      expect(success).toBe(true);
      expect(fetchRuntimeConfig).toHaveBeenCalledWith('http://192.168.1.50:8000/');

      const state = useServerStore.getState();
      expect(state.mode).toBe('self_host');
      expect(state.customApiUrl).toBe('http://192.168.1.50:8000/');
      expect(state.authUrl).toBe('https://auth.customer.com');
      expect(state.uploadUrl).toBe('https://upload.customer.com');
      expect(state.configured).toBe(true);
    });
  });

  describe('reset', () => {
    it('resets back to the initial state', async () => {
      (fetchRuntimeConfig as Mock).mockResolvedValue({
        authUrl: 'https://auth.customer.com',
        uploadUrl: 'https://upload.customer.com',
      });

      await useServerStore.getState().useSelfHost('http://192.168.1.50:8000');
      expect(useServerStore.getState().mode).toBe('self_host');

      useServerStore.getState().reset();
      const state = useServerStore.getState();
      expect(state.mode).toBe('hosted');
      expect(state.customApiUrl).toBeNull();
      expect(state.configured).toBe(false);
    });
  });
});
