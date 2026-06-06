import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: Record<string, unknown> | null;
  aboutContent: Record<string, unknown> | null;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: Record<string, unknown>[];
  news: Record<string, unknown>[];
}

interface ConfigActions {
  setWebsiteData: (data: Record<string, unknown>) => void;
  setAboutContent: (content: Record<string, unknown> | null) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: Record<string, unknown>[]) => void;
  setNews: (news: Record<string, unknown>[]) => void;
  reset: () => void;
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
    (set) => ({
      ...initialState,

      setWebsiteData: (websiteData) =>
        set((state) => {
          const current = state.websiteData ?? {};
          const nextSlug =
            (websiteData.slug as string | undefined) ?? (current.slug as string | undefined);

          return {
            websiteData: nextSlug ? { ...websiteData, slug: nextSlug } : websiteData,
          };
        }),

      setAboutContent: (aboutContent) => set({ aboutContent }),

      setMaintanance: (maintanance) => set({ maintanance }),

      setBusinessCount: (businessCount) => set({ businessCount }),

      setContactInfo: (contactInfo) => set({ contactInfo }),

      setFaqs: (faqs) => set({ faqs }),

      setNews: (news) => set({ news }),

      reset: () => set(initialState),
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
