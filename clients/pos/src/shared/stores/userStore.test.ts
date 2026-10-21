import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useUserStore } from './userStore';

const oauthClient = vi.hoisted(() => ({
  me: vi.fn(),
}));

const configClient = vi.hoisted(() => ({
  getConfig: vi.fn(),
}));

const settingsClient = vi.hoisted(() => ({
  getPOSSettings: vi.fn(),
  updatePOSSettings: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: vi.fn((client: { name?: string }) => {
    if (client.name === 'OauthServiceClient') return oauthClient;
    if (client.name === 'ConfigServiceClient') return configClient;
    if (client.name === 'SettingsServiceClient') return settingsClient;
    return {};
  }),
}));

const { getState, setState } = useUserStore;
const act = (fn: () => void) => {
  fn();
};

const initialState = {
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

beforeEach(() => {
  setState(initialState);
  vi.clearAllMocks();
  configClient.getConfig.mockResolvedValue({ maintenance: false, business: undefined });
});

describe('userStore', () => {
  it('starts with default unauthenticated state', () => {
    expect(getState().user).toBeNull();
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
    expect(getState().effectiveAccess).toBeNull();
    expect(getState().userId).toBeNull();
    expect(getState().addons.feedback).toBe(false);
    expect(getState().visits).toEqual([]);
  });

  it('sets token', () => {
    act(() => getState().setToken('abc123'));
    expect(getState().token).toBe('abc123');
  });

  it('sets user', () => {
    const user = { name: 'John', email: 'john@test.com' };
    act(() => getState().setUser(user));
    expect(getState().user).toEqual(user);
  });

  it('logs out and clears state', () => {
    setState({
      user: { name: 'John' },
      token: 'abc',
      isAuthenticated: true,
      effectiveAccess: null,
      userId: '123',
    });

    act(() => getState().logout());

    expect(getState().user).toBeNull();
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().effectiveAccess).toBeNull();
    expect(getState().userId).toBeNull();
  });

  it('clears error', () => {
    setState({ error: 'Something went wrong' });
    act(() => getState().clearError());
    expect(getState().error).toBeNull();
  });

  it('restores user info from generated oauth and config clients', async () => {
    oauthClient.me.mockResolvedValueOnce({
      userId: 'u1',
      businessId: 'b1',
      effectiveAccess: {
        businessId: 'b1',
        dashPermissions: [],
        subscribedApps: ['pos'],
        licenses: [],
        apps: [],
        isOwner: false,
        isAdmin: true,
        authorizationVersion: 1,
      },
    });
    configClient.getConfig.mockResolvedValueOnce({
      maintenance: false,
      business: {
        id: 'b1',
        name: 'Alice Store',
        type: 'retail',
        logoUrl: '',
        address: 'Main St',
        phone: '111',
        email: 'alice@test.com',
        website: '',
        timezone: 'UTC',
        currency: 'USD',
        country: 'US',
        plan: 'trial',
        createdAt: '',
        slug: 'alice-store',
      },
    });

    await getState().getUserInfo();

    expect(oauthClient.me).toHaveBeenCalledWith({});
    expect(configClient.getConfig).toHaveBeenCalledWith({});
    expect(getState().userId).toBe('u1');
    expect(getState().effectiveAccess?.isAdmin).toBe(true);
    expect(getState().isAuthenticated).toBe(true);
    expect(getState().businessProfile?.name).toBe('Alice Store');
    expect(getState().loading).toBe(false);
  });

  it('restores user info when config lookup fails', async () => {
    oauthClient.me.mockResolvedValueOnce({
      userId: '42',
      businessId: 'b1',
      effectiveAccess: {
        businessId: 'b1',
        dashPermissions: [],
        subscribedApps: ['pos'],
        licenses: [],
        apps: [],
        isOwner: false,
        isAdmin: true,
        authorizationVersion: 1,
      },
    });
    configClient.getConfig.mockRejectedValueOnce(new Error('config unavailable'));

    await getState().getUserInfo();

    expect(getState().userId).toBe('42');
    expect(getState().businessProfile).toBeNull();
    expect(getState().isAuthenticated).toBe(true);
  });

  it('sets error on getUserInfo failure', async () => {
    const error = new Error('Unauthorized') as Error & {
      response?: { data?: { message?: string } };
    };
    error.response = { data: { message: 'Unauthorized' } };
    oauthClient.me.mockRejectedValueOnce(error);

    await expect(getState().getUserInfo()).rejects.toBeTruthy();
    expect(getState().error).toBe('Unauthorized');
    expect(getState().loading).toBe(false);
  });

  it('uses generated error message on getUserInfo failure without response payload', async () => {
    oauthClient.me.mockRejectedValueOnce(new Error('Network'));

    await expect(getState().getUserInfo()).rejects.toBeTruthy();
    expect(getState().error).toBe('Network');
  });

  it('updates theme through generated POS settings', async () => {
    setState({ user: { name: 'Alice', theme: { mode: 'light' } } });
    settingsClient.getPOSSettings.mockResolvedValueOnce({
      settings: { businessProfile: '', features: undefined, receiptFooter: '' },
    });
    settingsClient.updatePOSSettings.mockResolvedValueOnce({
      settings: { theme: '{"primary":"#000"}' },
      message: 'ok',
    });

    await getState().setThemeAsync({ id: 'biz1', themeData: { primary: '#000' } });

    expect(settingsClient.updatePOSSettings).toHaveBeenCalledWith({
      businessProfile: '',
      theme: '{"primary":"#000"}',
      features: undefined,
      receiptFooter: '',
    });
    expect(getState().user?.theme).toEqual({ primary: '#000' });
    expect(getState().loading).toBe(false);
  });

  it('sets error on setThemeAsync failure', async () => {
    setState({ user: { name: 'Alice' } });
    settingsClient.getPOSSettings.mockRejectedValueOnce({
      response: { data: { message: 'Theme update failed' } },
    });

    await expect(getState().setThemeAsync({ id: 'biz1', themeData: {} })).rejects.toBeTruthy();
    expect(getState().error).toBe('Theme update failed');
    expect(getState().loading).toBe(false);
  });

  it('setUser replaces existing user', () => {
    act(() => getState().setUser({ name: 'A' }));
    act(() => getState().setUser({ name: 'B' }));
    expect(getState().user?.name).toBe('B');
  });
});
