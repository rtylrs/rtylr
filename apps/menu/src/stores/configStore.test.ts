import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useConfigStore } from './configStore';

const { getState, setState } = useConfigStore;

// Mock the api module
vi.mock('../shared/lib/api', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}));

beforeEach(() => {
  setState({
    websiteData: null,
    aboutContent: null,
    maintanance: false,
    businessCount: 0,
    contactInfo: { contactEmail: null, contactNumber: null },
    faqs: [],
    status: 'idle',
    error: null,
    news: [],
    loading: true,
  });
});

describe('configStore', () => {
  it('has correct initial state', () => {
    const s = getState();
    expect(s.websiteData).toBeNull();
    expect(s.aboutContent).toBeNull();
    expect(s.maintanance).toBe(false);
    expect(s.businessCount).toBe(0);
    expect(s.contactInfo).toEqual({ contactEmail: null, contactNumber: null });
    expect(s.faqs).toEqual([]);
    expect(s.status).toBe('idle');
    expect(s.error).toBeNull();
    expect(s.news).toEqual([]);
    expect(s.loading).toBe(true);
  });

  it('setWebsiteData updates websiteData', () => {
    getState().setWebsiteData({ title: 'My Site', logo: 'logo.png' });
    expect(getState().websiteData).toEqual({ title: 'My Site', logo: 'logo.png' });
  });

  it('setAboutContent updates aboutContent', () => {
    getState().setAboutContent('<p>About us</p>');
    expect(getState().aboutContent).toBe('<p>About us</p>');
  });

  it('setMaintanance toggles maintenance mode', () => {
    getState().setMaintanance(true);
    expect(getState().maintanance).toBe(true);
    getState().setMaintanance(false);
    expect(getState().maintanance).toBe(false);
  });

  it('setBusinessCount updates business count', () => {
    getState().setBusinessCount(42);
    expect(getState().businessCount).toBe(42);
  });

  it('setContactInfo updates contact info', () => {
    getState().setContactInfo({ contactEmail: 'hi@test.com', contactNumber: '+1234567890' });
    expect(getState().contactInfo.contactEmail).toBe('hi@test.com');
    expect(getState().contactInfo.contactNumber).toBe('+1234567890');
  });

  it('setFaqs replaces FAQs array', () => {
    const faqs = [{ q: 'What?', a: 'That.' }];
    getState().setFaqs(faqs);
    expect(getState().faqs).toEqual(faqs);
  });

  it('setNews replaces news array', () => {
    const news = [{ id: '1', title: 'Breaking' }];
    getState().setNews(news);
    expect(getState().news).toEqual(news);
  });

  it('setLoading updates loading state', () => {
    getState().setLoading(false);
    expect(getState().loading).toBe(false);
    getState().setLoading(true);
    expect(getState().loading).toBe(true);
  });

  it('fetchConfigData updates store on success', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      data: {
        config: {
          maintanance: true,
          aboutContent: 'About',
          count: 10,
          contactEmail: 'e@e.com',
          contactNumber: '555',
          faqs: [{ q: 'Q1', a: 'A1' }],
          news: [{ id: '1' }],
        },
      },
    });

    await getState().fetchConfigData();

    const s = getState();
    expect(s.maintanance).toBe(true);
    expect(s.aboutContent).toBe('About');
    expect(s.businessCount).toBe(10);
    expect(s.contactInfo.contactEmail).toBe('e@e.com');
    expect(s.faqs).toHaveLength(1);
    expect(s.loading).toBe(false);
  });

  it('fetchConfigData sets loading false on error', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    await getState().fetchConfigData();
    expect(getState().loading).toBe(false);
  });
});
