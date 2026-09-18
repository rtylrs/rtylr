import { create } from 'zustand';

interface ContactInfo {
  contactEmail: string | null;
  contactNumber: string | null;
}

interface ConfigState {
  websiteData: unknown;
  aboutContent: unknown;
  businessCount: number;
  contactInfo: ContactInfo;
  loading: boolean;
  error: string | null;

  // Actions
  setWebsiteData: (data: unknown) => void;
  setAboutContent: (content: unknown) => void;
  setBusinessCount: (count: number) => void;
  setContactInfo: (info: ContactInfo) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  websiteData: null,
  aboutContent: null,
  businessCount: 0,
  contactInfo: {
    contactEmail: null,
    contactNumber: null,
  },
  loading: false,
  error: null,
};

export const useConfigStore = create<ConfigState>((set) => ({
  ...initialState,

  setWebsiteData: (websiteData) => set({ websiteData }),
  setAboutContent: (aboutContent) => set({ aboutContent }),
  setBusinessCount: (businessCount) => set({ businessCount }),
  setContactInfo: (contactInfo) => set({ contactInfo }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  reset: () => set(initialState),
}));
