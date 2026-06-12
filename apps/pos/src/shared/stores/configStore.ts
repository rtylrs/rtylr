import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api/api';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: Record<string, unknown> | null;
  aboutContent: string | null;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: Record<string, unknown>[];
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;
  news: Record<string, unknown>[];
  loading: boolean;
}

interface ConfigActions {
  setWebsiteData: (data: Record<string, unknown>) => void;
  setAboutContent: (content: string | null) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: Record<string, unknown>[]) => void;
  setNews: (news: Record<string, unknown>[]) => void;
  setLoading: (loading: boolean) => void;
  setStatus: (status: 'idle' | 'loading' | 'succeeded' | 'failed') => void;
  setError: (error: string | null) => void;
  fetchConfigData: () => Promise<void>;
  updateAboutContent: (data: { aboutContent: string }) => Promise<void>;
  fetchFaqs: () => Promise<void>;
}

const initialState: ConfigState = {
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
};

export const useConfigStore = create<ConfigState & ConfigActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setWebsiteData: (websiteData) => set({ websiteData }),

      setAboutContent: (aboutContent) => set({ aboutContent }),

      setMaintanance: (maintanance) => set({ maintanance }),

      setBusinessCount: (businessCount) => set({ businessCount }),

      setContactInfo: (contactInfo) => set({ contactInfo }),

      setFaqs: (faqs) => set({ faqs }),

      setNews: (news) => set({ news }),

      setLoading: (loading) => set({ loading }),

      setStatus: (status) => set({ status }),

      setError: (error) => set({ error }),

      fetchConfigData: async () => {
        try {
          const response = await api.get('/config');
          const { maintanance, aboutContent, count, contactEmail, contactNumber, faqs, news } =
            response.data.config;

          set({
            aboutContent,
            maintanance,
            businessCount: count,
            contactInfo: {
              contactEmail,
              contactNumber,
            },
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
    }),
    {
      name: 'config-store',
      partialize: (state) => ({
        websiteData: state.websiteData,
        aboutContent: state.aboutContent,
        maintanance: state.maintanance,
        businessCount: state.businessCount,
        contactInfo: state.contactInfo,
        faqs: state.faqs,
        news: state.news,
      }),
    }
  )
);
