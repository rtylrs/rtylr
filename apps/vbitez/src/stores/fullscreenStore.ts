import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FullscreenState {
  isFullscreen: boolean;
}

interface FullscreenActions {
  setFullscreen: (isFullscreen: boolean) => void;
  toggleFullscreen: () => void;
}

const initialState: FullscreenState = {
  isFullscreen: true, // Default to fullscreen
};

export const useFullscreenStore = create<FullscreenState & FullscreenActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setFullscreen: (isFullscreen) => set({ isFullscreen }),

      toggleFullscreen: () => {
        const current = get().isFullscreen;
        set({ isFullscreen: !current });
      },
    }),
    {
      name: 'fullscreen-store',
    }
  )
);
