import { create } from 'zustand';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: any;
  aboutContent: any;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: any[];
  news: any[];
  loading: boolean;
  error: string | null;

  // Actions
  setWebsiteData: (data: any) => void;
  setAboutContent: (content: any) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: any[]) => void;
  setNews: (news: any[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  websiteData: null,
  aboutContent: null,
  maintanance: false,
  businessCount: 0,
  contactInfo: {
    contactEmail: null,
    contactNumber: null,
  },
  faqs: [],
  news: [],
  loading: false,
  error: null,
};

export const useConfigStore = create<ConfigState>((set) => ({
  ...initialState,

  setWebsiteData: (websiteData) => set({ websiteData }),
  setAboutContent: (aboutContent) => set({ aboutContent }),
  setMaintanance: (maintanance) => set({ maintanance }),
  setBusinessCount: (businessCount) => set({ businessCount }),
  setContactInfo: (contactInfo) => set({ contactInfo }),
  setFaqs: (faqs) => set({ faqs }),
  setNews: (news) => set({ news }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  reset: () => set(initialState),
}));