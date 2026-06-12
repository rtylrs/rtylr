import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useConfigStore } from './configStore';

// Mock the api module
vi.mock('../services/api/api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
  },
}));

import api from '../services/api/api';

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
  // ── Initial state ──────────────────────────────────────────────

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

  // ── Simple setters ─────────────────────────────────────────────

  it('setWebsiteData sets website data', () => {
    act(() => getState().setWebsiteData({ title: 'My Site' }));
    expect(getState().websiteData).toEqual({ title: 'My Site' });
  });

  it('setAboutContent sets about content', () => {
    act(() => getState().setAboutContent('About us text'));
    expect(getState().aboutContent).toBe('About us text');
  });

  it('setAboutContent accepts null', () => {
    act(() => getState().setAboutContent(null));
    expect(getState().aboutContent).toBeNull();
  });

  it('setMaintanance toggles maintenance mode', () => {
    act(() => getState().setMaintanance(true));
    expect(getState().maintanance).toBe(true);
    act(() => getState().setMaintanance(false));
    expect(getState().maintanance).toBe(false);
  });

  it('setBusinessCount sets count', () => {
    act(() => getState().setBusinessCount(42));
    expect(getState().businessCount).toBe(42);
  });

  it('setContactInfo sets contact info', () => {
    act(() => getState().setContactInfo({ contactEmail: 'a@b.com', contactNumber: '555-1234' }));
    expect(getState().contactInfo.contactEmail).toBe('a@b.com');
    expect(getState().contactInfo.contactNumber).toBe('555-1234');
  });

  it('setFaqs sets FAQ list', () => {
    const faqs = [{ q: 'What?', a: 'That.' }];
    act(() => getState().setFaqs(faqs));
    expect(getState().faqs).toEqual(faqs);
  });

  it('setNews sets news list', () => {
    act(() => getState().setNews([{ title: 'News 1' }]));
    expect(getState().news).toHaveLength(1);
  });

  it('setLoading sets loading state', () => {
    act(() => getState().setLoading(false));
    expect(getState().loading).toBe(false);
  });

  it('setStatus sets status', () => {
    act(() => getState().setStatus('loading'));
    expect(getState().status).toBe('loading');
  });

  it('setError sets error', () => {
    act(() => getState().setError('Something broke'));
    expect(getState().error).toBe('Something broke');
  });

  it('setError clears error with null', () => {
    setState({ error: 'old' });
    act(() => getState().setError(null));
    expect(getState().error).toBeNull();
  });

  // ── fetchConfigData ────────────────────────────────────────────

  it('fetches config data successfully', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: {
        config: {
          maintanance: true,
          aboutContent: 'About',
          count: 5,
          contactEmail: 'x@y.com',
          contactNumber: '111',
          faqs: [{ q: 'Q', a: 'A' }],
          news: [{ title: 'N' }],
        },
      },
    });

    await getState().fetchConfigData();

    expect(getState().maintanance).toBe(true);
    expect(getState().aboutContent).toBe('About');
    expect(getState().businessCount).toBe(5);
    expect(getState().contactInfo.contactEmail).toBe('x@y.com');
    expect(getState().faqs).toHaveLength(1);
    expect(getState().news).toHaveLength(1);
    expect(getState().loading).toBe(false);
  });

  it('sets loading false on fetchConfigData error', async () => {
    vi.mocked(api.get).mockRejectedValueOnce(new Error('fail'));
    await getState().fetchConfigData();
    expect(getState().loading).toBe(false);
  });

  // ── updateAboutContent ─────────────────────────────────────────

  it('updates about content successfully', async () => {
    vi.mocked(api.put).mockResolvedValueOnce({ data: { aboutContent: 'New about' } });
    await getState().updateAboutContent({ aboutContent: 'New about' });
    expect(getState().aboutContent).toBe('New about');
  });

  // ── fetchFaqs ──────────────────────────────────────────────────

  it('fetches FAQs successfully', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({ data: [{ q: 'Q', a: 'A' }] });
    await getState().fetchFaqs();
    expect(getState().faqs).toEqual([{ q: 'Q', a: 'A' }]);
  });
});
