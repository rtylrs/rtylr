import { create } from 'zustand';

interface ShortcutState {
  commandOpen: boolean;
  helpOpen: boolean;
}

interface ShortcutActions {
  toggleCommand: () => void;
  setCommand: (open: boolean) => void;
  toggleHelp: () => void;
  setHelp: (open: boolean) => void;
}

export const useShortcutStore = create<ShortcutState & ShortcutActions>()((set, get) => ({
  commandOpen: false,
  helpOpen: false,
  toggleCommand: () => set({ commandOpen: !get().commandOpen }),
  setCommand: (commandOpen) => set({ commandOpen }),
  toggleHelp: () => set({ helpOpen: !get().helpOpen }),
  setHelp: (helpOpen) => set({ helpOpen }),
}));
