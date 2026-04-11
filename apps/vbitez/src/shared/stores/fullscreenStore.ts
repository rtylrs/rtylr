import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FullscreenStore {
  isFullscreen: boolean;
  setFullscreen: (val: boolean) => void;
  toggleFullscreen: () => void;
}

export const useFullscreenStore = create<FullscreenStore>()(
  persist(
    (set) => ({
      isFullscreen: true,
      setFullscreen: (isFullscreen) => set({ isFullscreen }),
      toggleFullscreen: () => set((s) => ({ isFullscreen: !s.isFullscreen })),
    }),
    { name: 'pos-fullscreen' }
  )
);
