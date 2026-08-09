import { create } from 'zustand';
import api from '../shared/lib/api';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigStore {
  websiteData: unknown;
  aboutContent: unknown;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: unknown[];
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;
  news: unknown[];
  loading: boolean;
  setWebsiteData: (data: unknown) => void;
  setAboutContent: (content: unknown) => void;
  setMaintanance: (value: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: unknown[]) => void;
  setNews: (news: unknown[]) => void;
  setLoading: (loading: boolean) => void;
  fetchConfigData: () => Promise<void>;
  updateAboutContent: (data: { aboutContent: string }) => Promise<void>;
  fetchFaqs: () => Promise<void>;
}

export const useConfigStore = create<ConfigStore>()((set) => ({
  websiteData: null,
  aboutContent: null,
  maintanance: false,
  businessCount: 0,
  contactInfo: {
    contactEmail: null,
    contactNumber: null,
  },
  faqs: [],
  status: 'idle',
  error: null,
  news: [],
  loading: true,

  setWebsiteData: (websiteData) => set({ websiteData }),
  setAboutContent: (aboutContent) => set({ aboutContent }),
  setMaintanance: (maintanance) => set({ maintanance }),
  setBusinessCount: (businessCount) => set({ businessCount }),
  setContactInfo: (contactInfo) => set({ contactInfo }),
  setFaqs: (faqs) => set({ faqs }),
  setNews: (news) => set({ news }),
  setLoading: (loading) => set({ loading }),

  fetchConfigData: async () => {
    try {
      const response = await api.get('/config');
      const { maintanance, aboutContent, count, contactEmail, contactNumber, faqs, news } =
        response.data.config;
      set({
        aboutContent,
        maintanance,
        businessCount: count,
        contactInfo: { contactEmail, contactNumber },
        faqs,
        news,
        loading: false,
      });
    } catch (error) {
      set({ loading: false });
    }
  },

  updateAboutContent: async (data) => {
    try {
      const response = await api.put('/config/update', data);
      set({ aboutContent: response.data.aboutContent });
    } catch (error) {}
  },

  fetchFaqs: async () => {
    try {
      const response = await api.get('/config/faqs');
      set({ faqs: response.data });
    } catch (error) {}
  },
}));
