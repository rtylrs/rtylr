import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

export interface ConfigState {
  websiteData: any;
  aboutContent: any;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: any[];
  news: any[];
  loading: boolean;
}

interface ConfigActions {
  setWebsiteData: (data: any) => void;
  setAboutContent: (content: any) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (contactInfo: ContactInfo) => void;
  setFaqs: (faqs: any[]) => void;
  setNews: (news: any[]) => void;
  setLoading: (loading: boolean) => void;
}

type ConfigStore = ConfigState & ConfigActions;

export const useConfigStore = create<ConfigStore>()(
  persist(
    (set) => ({
      // State
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
      loading: true,

      // Actions
      setWebsiteData: (websiteData: any) => set({ websiteData }),
      setAboutContent: (aboutContent: any) => set({ aboutContent }),
      setMaintanance: (maintanance: boolean) => set({ maintanance }),
      setBusinessCount: (businessCount: number) => set({ businessCount }),
      setContactInfo: (contactInfo: ContactInfo) => set({ contactInfo }),
      setFaqs: (faqs: any[]) => set({ faqs }),
      setNews: (news: any[]) => set({ news }),
      setLoading: (loading: boolean) => set({ loading }),
    }),
    {
      name: 'rtylr:voxire:config',
    }
  )
);