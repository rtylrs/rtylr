import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface SidebarState {
  sectionsVisible: Record<string, boolean>;
}

interface SidebarActions {
  toggleSection: (section: string) => void;
  initializeSections: (sections: string[]) => void;
}

export const useSidebarStore = create<SidebarState & SidebarActions>()(
  persist(
    (set, get) => ({
      sectionsVisible: {},

      toggleSection: (section: string) => {
        const { sectionsVisible } = get();
        set({
          sectionsVisible: {
            ...sectionsVisible,
            [section]: !sectionsVisible[section],
          },
        });
      },

      initializeSections: (sections: string[]) => {
        const { sectionsVisible } = get();
        const newSectionsVisible = { ...sectionsVisible };

        sections.forEach((section) => {
          if (!(section in newSectionsVisible)) {
            newSectionsVisible[section] = false;
          }
        });

        set({ sectionsVisible: newSectionsVisible });
      },
    }),
    {
      name: 'rtylr:sidebar',
      storage: createJSONStorage(() => localStorage),
    }
  )
);
