import { create } from 'zustand';

export type CaretView = 'chat' | 'history' | 'settings';

interface UIState {
  view: CaretView;
  composerExpanded: boolean;
  pendingApprovalId: string | null;
}

interface UIActions {
  setView: (v: CaretView) => void;
  toggleComposer: () => void;
  setComposerExpanded: (v: boolean) => void;
  setPendingApproval: (id: string | null) => void;
}

export type UIStore = UIState & UIActions;

export const useCaretUIStore = create<UIStore>((set) => ({
  view: 'chat',
  composerExpanded: false,
  pendingApprovalId: null,
  setView: (view) => set({ view }),
  toggleComposer: () => set((s) => ({ composerExpanded: !s.composerExpanded })),
  setComposerExpanded: (composerExpanded) => set({ composerExpanded }),
  setPendingApproval: (pendingApprovalId) => set({ pendingApprovalId }),
}));
