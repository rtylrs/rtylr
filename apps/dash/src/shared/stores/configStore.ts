import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: Record<string, unknown> | null;
  aboutContent: Record<string, unknown> | null;
  maintenance: boolean;
  maintanance: boolean;
  businessCount: number;
  contactInfo: ContactInfo;
  faqs: Record<string, unknown>[];
  news: Record<string, unknown>[];
}

interface ConfigActions {
  setWebsiteData: (data: Record<string, unknown>) => void;
  setAboutContent: (content: Record<string, unknown> | null) => void;
  setMaintenance: (maintenance: boolean) => void;
  setMaintanance: (maintanance: boolean) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setFaqs: (faqs: Record<string, unknown>[]) => void;
  setNews: (news: Record<string, unknown>[]) => void;
  reset: () => void;
}

const configKeyAliases: Record<string, string> = {
  owner_id: 'ownerId',
  site_name: 'siteName',
  logo_url: 'logoUrl',
  favicon_url: 'faviconUrl',
  meta_tags: 'metaTags',
  contact_email: 'contactEmail',
  contact_number: 'contactNumber',
  sender_email: 'senderEmail',
  maps_url: 'mapsUrl',
  maintenance_mode: 'maintenanceMode',
  maintenance_notice: 'maintenanceNotice',
  addons_enabled: 'addonsEnabled',
  enable_reviews: 'enableReviews',
  enable_wishlist: 'enableWishlist',
  enable_newsletter_signup: 'enableNewsletterSignup',
  about_content: 'aboutContent',
  footer_sub_text: 'footerSubText',
};

export const normalizeConfigData = (data: Record<string, unknown>): Record<string, unknown> => {
  const normalized = { ...data };

  Object.entries(configKeyAliases).forEach(([sourceKey, targetKey]) => {
    if (normalized[targetKey] === undefined && normalized[sourceKey] !== undefined) {
      normalized[targetKey] = normalized[sourceKey];
    }
  });

  return normalized;
};

const initialState: ConfigState = {
  websiteData: null,
  aboutContent: null,
  maintenance: false,
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
          const normalizedWebsiteData = normalizeConfigData(websiteData);
          const current = state.websiteData ?? {};
          const nextSlug =
            (normalizedWebsiteData.slug as string | undefined) ??
            (current.slug as string | undefined);

          return {
            websiteData: nextSlug
              ? { ...normalizedWebsiteData, slug: nextSlug }
              : normalizedWebsiteData,
          };
        }),

      setAboutContent: (aboutContent) => set({ aboutContent }),

      setMaintenance: (maintenance) => set({ maintenance, maintanance: maintenance }),

      setMaintanance: (maintanance) => set({ maintenance: maintanance, maintanance }),

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
        maintenance: state.maintenance,
        maintanance: state.maintanance,
        businessCount: state.businessCount,
        contactInfo: state.contactInfo,
        faqs: state.faqs,
        news: state.news,
      }),
    }
  )
);
