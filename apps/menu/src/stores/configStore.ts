import { create } from 'zustand';
import api from '../lib/api';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigStore {
  websiteData: any;
  aboutContent: any;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: any[];
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;
  news: any[];
  loading: boolean;
  setWebsiteData: (data: any) => void;
  setAboutContent: (content: any) => void;
  setMaintanance: (value: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: any[]) => void;
  setNews: (news: any[]) => void;
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
      const {
        maintanance,
        aboutContent,
        count,
        contactEmail,
        contactNumber,
        faqs,
        news,
      } = response.data.config;
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
      console.error('Error fetching config data:', error);
    }
  },

  updateAboutContent: async (data) => {
    try {
      const response = await api.put('/config/update', data);
      set({ aboutContent: response.data.aboutContent });
    } catch (error) {
      console.error('Error updating about content:', error);
    }
  },

  fetchFaqs: async () => {
    try {
      const response = await api.get('/config/faqs');
      set({ faqs: response.data });
    } catch (error) {
      console.error('Error fetching FAQs:', error);
    }
  },
}));
