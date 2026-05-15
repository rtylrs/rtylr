import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { DEFAULT_MODE, type CaretMode } from '../types/mode';
import { DEFAULT_MODEL, type CaretModel } from '../types/model';

interface CaretSettingsState {
  mode: CaretMode;
  model: CaretModel;
  voiceEnabled: boolean;
  autoApproveReadOnly: boolean;
  showThinking: boolean;
  panelWidth: number;
}

interface CaretSettingsActions {
  setMode: (mode: CaretMode) => void;
  setModel: (model: CaretModel) => void;
  setVoiceEnabled: (enabled: boolean) => void;
  setAutoApproveReadOnly: (auto: boolean) => void;
  setShowThinking: (show: boolean) => void;
  setPanelWidth: (px: number) => void;
}

export type CaretSettingsStore = CaretSettingsState & CaretSettingsActions;

const initial: CaretSettingsState = {
  mode: DEFAULT_MODE,
  model: DEFAULT_MODEL,
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
      setModel: (model) => set({ model }),
      setVoiceEnabled: (voiceEnabled) => set({ voiceEnabled }),
      setAutoApproveReadOnly: (autoApproveReadOnly) => set({ autoApproveReadOnly }),
      setShowThinking: (showThinking) => set({ showThinking }),
      setPanelWidth: (panelWidth) => set({ panelWidth: Math.max(320, Math.min(800, panelWidth)) }),
    }),
    { name: 'caret-settings' }
  )
);
