import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface MenuItem {
  _id: string;
  itemId?: string;
  name: string;
  type?: 'item';
  details?: {
    name: string;
    description?: string;
    price?: number;
    sku?: string;
    isActive?: boolean;
  };
  [key: string]: unknown;
}

interface Category {
  _id: string;
  name: string;
  type: 'category';
  menuItems: MenuItem[];
  [key: string]: unknown;
}

interface Banner {
  _id: string;
  bannerImage: string;
  menuId: string;
}

interface Menu {
  _id: string;
  name?: string;
  itemCount?: number;
  isActive?: boolean;
  items: (Category | Banner | MenuItem)[];
}

interface MenuState {
  id: string | null;
  items: (Category | Banner | MenuItem)[];
  allMenus: Menu[];
  activeMenuId: string | null;
  status: 'idle' | 'loading' | 'success' | 'error';
}

interface MenuActions {
  setId: (id: string | null) => void;
  setItems: (items: (Category | Banner | MenuItem)[]) => void;
  setAllMenus: (menus: Menu[]) => void;
  setActiveMenuId: (id: string | null) => void;
  fetchMenu: (userId: string, force?: boolean) => Promise<void>;
  addCategory: (category: Category) => void;
  editCategory: (id: string, updates: Partial<Category>) => void;
  deleteCategory: (id: string) => void;
  addBanner: (banner: Banner) => void;
}

function isCategory(item: Category | Banner | MenuItem): item is Category {
  return (item as Category).type === 'category';
}

const initialState: MenuState = {
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
  status: 'idle',
};

export const useMenuStore = create<MenuState & MenuActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setId: (id) => set({ id }),

      setItems: (items) => set({ items }),

      setAllMenus: (allMenus) => set({ allMenus }),

      setActiveMenuId: (activeMenuId) => set({ activeMenuId }),

      fetchMenu: async (userId, force) => {
        set({ status: 'loading' });
        try {
          // Placeholder - actual API call would go here
          set({ status: 'success' });
        } catch {
          set({ status: 'error' });
        }
      },

      addCategory: (category) => set((state) => ({ items: [...state.items, category] })),

      editCategory: (id, updates) =>
        set((state) => ({
          items: state.items.map((item) =>
            isCategory(item) && item.id === id ? { ...item, ...updates } : item
          ),
        })),

      deleteCategory: (id) =>
        set((state) => ({
          items: state.items.filter((item) => !isCategory(item) || item.id !== id),
        })),

      addBanner: (banner) => set((state) => ({ items: [...state.items, banner] })),
    }),
    {
      name: 'menu-store',
      partialize: (state) => ({
        id: state.id,
        items: state.items,
        allMenus: state.allMenus,
        activeMenuId: state.activeMenuId,
      }),
    }
  )
);
