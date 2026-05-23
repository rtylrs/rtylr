import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type WindowType =
  | 'pos'
  | 'erp'
  | 'crm'
  | 'hr'
  | 'finance'
  | 'flow'
  | 'settings'
  | 'about'
  | 'docs'
  | 'pricing'
  | 'landing';
export type Experience = 'os' | 'website';
export type ColorMode = 'system' | 'light' | 'dark';
export type Wallpaper = 'midnight' | 'aurora' | 'dusk' | 'slate' | 'forest';

export interface WindowState {
  id: string;
  type: WindowType;
  title: string;
  url?: string;
  x: number;
  y: number;
  width: number;
  height: number;
  zIndex: number;
  minimized: boolean;
}

interface DesktopState {
  windows: WindowState[];
  globalZ: number;
  iconPositions: Record<string, { x: number; y: number }>;
  experience: Experience;
  colorMode: ColorMode;
  wallpaper: Wallpaper;
  animationsEnabled: boolean;

  addWindow: (type: WindowType, url?: string) => void;
  closeWindow: (id: string) => void;
  bringToFront: (id: string) => void;
  minimizeWindow: (id: string) => void;
  restoreWindow: (id: string) => void;
  updateWindowPosition: (id: string, x: number, y: number) => void;
  updateWindowSize: (id: string, width: number, height: number) => void;
  closeAllWindows: () => void;
  setExperience: (exp: Experience) => void;
  setColorMode: (mode: ColorMode) => void;
  setWallpaper: (wallpaper: Wallpaper) => void;
  setAnimationsEnabled: (enabled: boolean) => void;
  setIconPosition: (key: string, pos: { x: number; y: number }) => void;
}

const WINDOW_DEFAULTS: Record<WindowType, { title: string; width: number; height: number }> = {
  pos: { title: 'POS · vbitez', width: 720, height: 460 },
  erp: { title: 'ERP', width: 680, height: 440 },
  crm: { title: 'CRM', width: 680, height: 440 },
  hr: { title: 'HR', width: 640, height: 420 },
  finance: { title: 'Finance', width: 660, height: 440 },
  flow: { title: 'Flow', width: 640, height: 420 },
  settings: { title: 'Display Options', width: 480, height: 400 },
  about: { title: 'About rtylr', width: 440, height: 360 },
  docs: { title: 'Docs', width: 640, height: 480 },
  pricing: { title: 'Pricing', width: 700, height: 520 },
  landing: { title: 'rtylr.com', width: 1100, height: 660 },
};

let idCounter = 0;

export const useDesktopStore = create<DesktopState>()(
  persist(
    (set, get) => ({
      windows: [],
      globalZ: 600,
      iconPositions: {},
      experience: 'website',
      colorMode: 'system',
      wallpaper: 'midnight',
      animationsEnabled: true,

      addWindow: (type, url?) => {
        const existing = get().windows.find((w) => w.type === type);
        if (existing) {
          if (existing.minimized) {
            set((s) => ({
              globalZ: s.globalZ + 1,
              windows: s.windows.map((w) =>
                w.id === existing.id ? { ...w, minimized: false, zIndex: s.globalZ + 1 } : w
              ),
            }));
          } else {
            get().bringToFront(existing.id);
          }
          return;
        }
        const defaults = WINDOW_DEFAULTS[type];
        const z = get().globalZ + 1;
        const offset = (idCounter % 8) * 24;
        const id = `${type}-${++idCounter}`;

        // Center on screen for landing window; cascade for others
        const vw = typeof window !== 'undefined' ? window.innerWidth : 1280;
        const vh = typeof window !== 'undefined' ? window.innerHeight : 800;
        const menuBarH = 36;
        const x =
          type === 'landing' ? Math.max(0, Math.floor((vw - defaults.width) / 2)) : 80 + offset;
        const y =
          type === 'landing'
            ? Math.max(0, Math.floor((vh - menuBarH - defaults.height) / 2))
            : 40 + offset;

        set((s) => ({
          globalZ: z,
          windows: [
            ...s.windows,
            {
              id,
              type,
              title: defaults.title,
              url,
              x,
              y,
              width: defaults.width,
              height: defaults.height,
              zIndex: z,
              minimized: false,
            },
          ],
        }));
      },

      closeWindow: (id) => set((s) => ({ windows: s.windows.filter((w) => w.id !== id) })),

      bringToFront: (id) =>
        set((s) => ({
          globalZ: s.globalZ + 1,
          windows: s.windows.map((w) => (w.id === id ? { ...w, zIndex: s.globalZ + 1 } : w)),
        })),

      minimizeWindow: (id) =>
        set((s) => ({
          windows: s.windows.map((w) => (w.id === id ? { ...w, minimized: true } : w)),
        })),

      restoreWindow: (id) =>
        set((s) => ({
          globalZ: s.globalZ + 1,
          windows: s.windows.map((w) =>
            w.id === id ? { ...w, minimized: false, zIndex: s.globalZ + 1 } : w
          ),
        })),

      updateWindowPosition: (id, x, y) =>
        set((s) => ({ windows: s.windows.map((w) => (w.id === id ? { ...w, x, y } : w)) })),

      updateWindowSize: (id, width, height) =>
        set((s) => ({
          windows: s.windows.map((w) => (w.id === id ? { ...w, width, height } : w)),
        })),

      closeAllWindows: () => set({ windows: [] }),

      setExperience: (experience) => set({ experience }),
      setColorMode: (colorMode) => set({ colorMode }),
      setWallpaper: (wallpaper) => set({ wallpaper }),
      setAnimationsEnabled: (animationsEnabled) => set({ animationsEnabled }),
      setIconPosition: (key, pos) =>
        set((s) => ({ iconPositions: { ...s.iconPositions, [key]: pos } })),
    }),
    {
      name: 'rtylr-desktop',
      storage: createJSONStorage(() => localStorage),
      partialize: (s) => ({
        experience: s.experience,
        colorMode: s.colorMode,
        wallpaper: s.wallpaper,
        animationsEnabled: s.animationsEnabled,
        iconPositions: s.iconPositions,
      }),
    }
  )
);
