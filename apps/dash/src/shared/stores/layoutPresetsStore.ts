import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { tenantStorage } from '../utils/tenantStorage';
import type { MenuLayoutConfig } from '@voxire/menu';
import { DEFAULT_LAYOUT_CONFIG } from '@voxire/menu';

export interface LayoutPreset {
  id: string;
  name: string;
  config: MenuLayoutConfig;
  createdAt: number;
}

interface LayoutPresetsState {
  presets: LayoutPreset[];
  activePresetId: string | null;
  // Live draft config (what the user is currently editing)
  draftConfig: MenuLayoutConfig;
  // Actions
  createPreset: (name: string, config?: MenuLayoutConfig) => string;
  updatePreset: (id: string, updates: Partial<Omit<LayoutPreset, 'id' | 'createdAt'>>) => void;
  deletePreset: (id: string) => void;
  activatePreset: (id: string) => void;
  setDraftConfig: (config: MenuLayoutConfig) => void;
  patchDraftConfig: (patch: Partial<MenuLayoutConfig>) => void;
  saveActivePreset: () => void;
}

const DEFAULT_PRESET_ID = 'default';

const initialPresets: LayoutPreset[] = [
  {
    id: DEFAULT_PRESET_ID,
    name: 'Default',
    config: DEFAULT_LAYOUT_CONFIG,
    createdAt: Date.now(),
  },
];

export const useLayoutPresetsStore = create<LayoutPresetsState>()(
  persist(
    (set, get) => ({
      presets: initialPresets,
      activePresetId: DEFAULT_PRESET_ID,
      draftConfig: DEFAULT_LAYOUT_CONFIG,

      createPreset: (name, config) => {
        const id = `preset_${Date.now()}`;
        const newPreset: LayoutPreset = {
          id,
          name,
          config: config ?? get().draftConfig,
          createdAt: Date.now(),
        };
        set((s) => ({ presets: [...s.presets, newPreset] }));
        return id;
      },

      updatePreset: (id, updates) => {
        set((s) => ({
          presets: s.presets.map((p) => (p.id === id ? { ...p, ...updates } : p)),
        }));
      },

      deletePreset: (id) => {
        set((s) => {
          const remaining = s.presets.filter((p) => p.id !== id);
          const newActive = s.activePresetId === id ? (remaining[0]?.id ?? null) : s.activePresetId;
          const activePreset = remaining.find((p) => p.id === newActive);
          return {
            presets: remaining,
            activePresetId: newActive,
            draftConfig: activePreset?.config ?? DEFAULT_LAYOUT_CONFIG,
          };
        });
      },

      activatePreset: (id) => {
        const preset = get().presets.find((p) => p.id === id);
        if (!preset) return;
        set({ activePresetId: id, draftConfig: preset.config });
      },

      setDraftConfig: (config) => {
        set({ draftConfig: config });
      },

      patchDraftConfig: (patch) => {
        set((s) => ({ draftConfig: { ...s.draftConfig, ...patch } }));
      },

      saveActivePreset: () => {
        const { activePresetId, draftConfig, presets } = get();
        if (!activePresetId) return;
        set({
          presets: presets.map((p) =>
            p.id === activePresetId ? { ...p, config: draftConfig } : p
          ),
        });
      },
    }),
    {
      name: 'layout-presets',
      storage: tenantStorage,
    }
  )
);
