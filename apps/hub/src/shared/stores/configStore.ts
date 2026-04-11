import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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
}

interface ConfigActions {
  setWebsiteData: (data: any) => void;
  setAboutContent: (content: any | null) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: any[]) => void;
  setNews: (news: any[]) => void;
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
  news: [],
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
