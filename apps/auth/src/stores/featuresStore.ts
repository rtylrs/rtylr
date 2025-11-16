import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface FeatureFlags {
  advanced_search: boolean;
  beta_features: boolean;
  dark_mode: boolean;
  google_signup: boolean;
}

export interface FeaturesState {
  flags: FeatureFlags;
  loading: boolean;
}

interface FeaturesActions {
  setFlags: (flags: FeatureFlags) => void;
  setLoading: (loading: boolean) => void;
}

type FeaturesStore = FeaturesState & FeaturesActions;

const defaultFlags: FeatureFlags = {
  advanced_search: false,
  beta_features: false,
  dark_mode: false,
  google_signup: false,
};

export const useFeaturesStore = create<FeaturesStore>()(
  persist(
    (set) => ({
      flags: defaultFlags,
      loading: true,

      setFlags: (flags: FeatureFlags) => set({ flags }),
      setLoading: (loading: boolean) => set({ loading }),
    }),
    {
      name: 'rtylr:voxire:features',
    }
  )
);
