import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface MenuItem {
  _id: string;
  name: string;
  [key: string]: any;
}

interface Category {
  _id: string;
  name: string;
  type: 'category';
  menuItems: MenuItem[];
  [key: string]: any;
}

interface Banner {
  _id: string;
  bannerImage: string;
  menuId: string;
}

interface Menu {
  _id: string;
  items: (Category | Banner)[];
}

interface MenuState {
  id: string | null;
  items: (Category | Banner)[];
  allMenus: Menu[];
  activeMenuId: string | null;
}

interface MenuActions {
  setId: (id: string | null) => void;
  setItems: (items: (Category | Banner)[]) => void;
  setAllMenus: (menus: Menu[]) => void;
  setActiveMenuId: (id: string | null) => void;
}

function isCategory(item: Category | Banner): item is Category {
  return (item as Category).type === 'category';
}

const initialState: MenuState = {
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
};

export const useMenuStore = create<MenuState & MenuActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setId: (id) => set({ id }),

      setItems: (items) => set({ items }),

      setAllMenus: (allMenus) => set({ allMenus }),

      setActiveMenuId: (activeMenuId) => set({ activeMenuId }),
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