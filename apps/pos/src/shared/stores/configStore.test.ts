import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useConfigStore } from './configStore';

const configClient = vi.hoisted(() => ({
  getConfig: vi.fn(),
  updateConfig: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => configClient,
}));

const { getState, setState } = useConfigStore;
const act = (fn: () => void) => {
  fn();
};

const initialState = {
  websiteData: null,
  aboutContent: null,
  maintanance: false,
  businessCount: 0,
  contactInfo: { contactEmail: null, contactNumber: null },
  faqs: [],
  status: 'idle' as const,
  error: null,
  news: [],
  loading: true,
};

beforeEach(() => {
  setState(initialState);
  vi.clearAllMocks();
});

describe('configStore', () => {
  it('starts with default state', () => {
    expect(getState().websiteData).toBeNull();
    expect(getState().aboutContent).toBeNull();
    expect(getState().maintanance).toBe(false);
    expect(getState().businessCount).toBe(0);
    expect(getState().contactInfo).toEqual({ contactEmail: null, contactNumber: null });
    expect(getState().faqs).toEqual([]);
    expect(getState().status).toBe('idle');
    expect(getState().error).toBeNull();
    expect(getState().news).toEqual([]);
    expect(getState().loading).toBe(true);
  });

  it('setters update local config state', () => {
    act(() => getState().setWebsiteData({ title: 'My Site' }));
    act(() => getState().setAboutContent('About us text'));
    act(() => getState().setMaintanance(true));
    act(() => getState().setBusinessCount(42));
    act(() => getState().setContactInfo({ contactEmail: 'a@b.com', contactNumber: '555-1234' }));
    act(() => getState().setFaqs([{ q: 'What?', a: 'That.' }]));
    act(() => getState().setNews([{ title: 'News 1' }]));
    act(() => getState().setStatus('loading'));
    act(() => getState().setError('Something broke'));

    expect(getState().websiteData).toEqual({ title: 'My Site' });
    expect(getState().aboutContent).toBe('About us text');
    expect(getState().maintanance).toBe(true);
    expect(getState().businessCount).toBe(42);
    expect(getState().contactInfo.contactEmail).toBe('a@b.com');
    expect(getState().faqs).toHaveLength(1);
    expect(getState().news).toHaveLength(1);
    expect(getState().status).toBe('loading');
    expect(getState().error).toBe('Something broke');
  });

  it('fetches config data through generated hub config client', async () => {
    configClient.getConfig.mockResolvedValueOnce({
      maintenance: true,
      business: {
        id: 'b1',
        name: 'Store',
        email: 'x@y.com',
        phone: '111',
      },
    });

    await getState().fetchConfigData();

    expect(configClient.getConfig).toHaveBeenCalledWith({});
    expect(getState().maintanance).toBe(true);
    expect(getState().businessCount).toBe(1);
    expect(getState().contactInfo.contactEmail).toBe('x@y.com');
    expect(getState().faqs).toEqual([]);
    expect(getState().news).toEqual([]);
    expect(getState().loading).toBe(false);
  });

  it('sets loading false on fetchConfigData error', async () => {
    configClient.getConfig.mockRejectedValueOnce(new Error('fail'));
    await getState().fetchConfigData();
    expect(getState().loading).toBe(false);
  });

  it('updates about content through generated hub config client', async () => {
    configClient.updateConfig.mockResolvedValueOnce({
      message: 'ok',
      config: { aboutContent: 'New about' },
    });

    await getState().updateAboutContent({ aboutContent: 'New about' });

    expect(configClient.updateConfig).toHaveBeenCalledWith(expect.objectContaining({
      aboutContent: 'New about',
    }));
    expect(getState().aboutContent).toBe('New about');
  });

  it('keeps FAQ fetch explicit until a generated FAQ contract exists', async () => {
    await getState().fetchFaqs();
    expect(getState().faqs).toEqual([]);
  });
});
