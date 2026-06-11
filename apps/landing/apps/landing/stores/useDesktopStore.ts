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
  | 'landing'
  | 'shortcuts'
  | 'terminal'
  | 'doom'
  | 'status'
  | 'snake';
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
  experience: Experience;
  colorMode: ColorMode;
  wallpaper: Wallpaper;
  animationsEnabled: boolean;
  spotlightOpen: boolean;

  addWindow: (type: WindowType, url?: string) => void;
  setSpotlightOpen: (v: boolean) => void;
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
}

const WINDOW_DEFAULTS: Record<WindowType, { title: string; width: number; height: number }> = {
  pos: { title: 'POS', width: 860, height: 560 },
  erp: { title: 'ERP', width: 860, height: 560 },
  crm: { title: 'CRM', width: 860, height: 560 },
  hr: { title: 'HR', width: 860, height: 560 },
  finance: { title: 'Finance', width: 860, height: 560 },
  flow: { title: 'Flow', width: 860, height: 560 },
  settings: { title: 'Display Options', width: 500, height: 420 },
  about: { title: 'About rtylr', width: 900, height: 600 },
  docs: { title: 'Docs', width: 860, height: 580 },
  pricing: { title: 'Pricing', width: 860, height: 580 },
  landing: { title: 'rtylr.com', width: 1100, height: 760 },
  shortcuts: { title: 'Keyboard Shortcuts', width: 480, height: 460 },
  terminal: { title: 'Terminal', width: 700, height: 460 },
  doom: { title: 'Doom', width: 860, height: 580 },
  status: { title: 'System Status', width: 860, height: 560 },
  snake: { title: 'Snake', width: 520, height: 520 },
};

let idCounter = 0;

export const useDesktopStore = create<DesktopState>()(
  persist(
    (set, get) => ({
      windows: [],
      globalZ: 600,
      experience: 'os',
      colorMode: 'system',
      wallpaper: 'slate',
      animationsEnabled: true,
      spotlightOpen: false,

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
        const width = Math.min(defaults.width, vw - 16);
        const height = Math.min(defaults.height, vh - menuBarH - 16);
        const x =
          type === 'landing'
            ? Math.max(0, Math.floor((vw - width) / 2))
            : Math.max(0, Math.min(80 + offset, vw - width - 8));
        const y =
          type === 'landing'
            ? Math.max(0, Math.floor((vh - menuBarH - height) / 2))
            : Math.max(0, Math.min(40 + offset, vh - menuBarH - height - 8));

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
              width,
              height,
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
      setSpotlightOpen: (spotlightOpen) => set({ spotlightOpen }),
    }),
    {
      name: 'rtylr-desktop',
      version: 2,
      storage: createJSONStorage(() => localStorage),
      partialize: (s) => ({
        experience: s.experience,
        colorMode: s.colorMode,
        wallpaper: s.wallpaper,
        animationsEnabled: s.animationsEnabled,
      }),
    }
  )
);
