import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { DEFAULT_MODE, type CaretMode } from '../types/mode';

// The model is chosen internally by the backend (DeepSeek / NVIDIA NIM) — users
// don't pick it, so there is no model field here anymore.
interface CaretSettingsState {
  mode: CaretMode;
  voiceEnabled: boolean;
  autoApproveReadOnly: boolean;
  showThinking: boolean;
  panelWidth: number;
}

interface CaretSettingsActions {
  setMode: (mode: CaretMode) => void;
  setVoiceEnabled: (enabled: boolean) => void;
  setAutoApproveReadOnly: (auto: boolean) => void;
  setShowThinking: (show: boolean) => void;
  setPanelWidth: (px: number) => void;
}

export type CaretSettingsStore = CaretSettingsState & CaretSettingsActions;

const initial: CaretSettingsState = {
  mode: DEFAULT_MODE,
  voiceEnabled: true,
  autoApproveReadOnly: true,
  showThinking: true,
  panelWidth: 380,
};

export const useCaretSettingsStore = create<CaretSettingsStore>()(
  persist(
    (set) => ({
      ...initial,
      setMode: (mode) => set({ mode }),
      setVoiceEnabled: (voiceEnabled) => set({ voiceEnabled }),
      setAutoApproveReadOnly: (autoApproveReadOnly) => set({ autoApproveReadOnly }),
      setShowThinking: (showThinking) => set({ showThinking }),
      setPanelWidth: (panelWidth) => set({ panelWidth: Math.max(320, Math.min(800, panelWidth)) }),
    }),
    { name: 'caret-settings' }
  )
);
