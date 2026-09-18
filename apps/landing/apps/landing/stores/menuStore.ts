import { create } from 'zustand';

interface MenuItem {
  _id: string;
  name: string;
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
  items: (Category | Banner)[];
}

interface MenuState {
  id: string | null;
  items: (Category | Banner)[];
  allMenus: Menu[];
  activeMenuId: string | null;
  loading: boolean;
  error: string | null;

  // Actions
  setId: (id: string | null) => void;
  setItems: (items: (Category | Banner)[]) => void;
  setAllMenus: (menus: Menu[]) => void;
  setActiveMenuId: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  // Menu item actions
  addMenuItem: (categoryId: string, item: MenuItem) => void;
  editMenuItem: (itemId: string, updatedItem: MenuItem) => void;
  deleteMenuItem: (categoryId: string, itemId: string) => void;

  // Category actions
  addCategory: (category: Category) => void;
  editCategory: (categoryId: string, updatedCategory: Category) => void;
  deleteCategory: (categoryId: string) => void;
  renameCategory: (categoryId: string, newName: string) => void;

  // Banner actions
  addBanner: (banner: Banner) => void;

  reset: () => void;
}

function isCategory(item: Category | Banner): item is Category {
  return (item as Category).type === 'category';
}

const initialState = {
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
  loading: false,
  error: null,
};

export const useMenuStore = create<MenuState>((set, get) => ({
  ...initialState,

  setId: (id) => set({ id }),
  setItems: (items) => set({ items }),
  setAllMenus: (allMenus) => set({ allMenus }),
  setActiveMenuId: (activeMenuId) => set({ activeMenuId, id: activeMenuId, items: [] }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  addMenuItem: (categoryId, item) =>
    set((state) => {
      const newItems = [...state.items];
      const categoryIndex = newItems.findIndex((cat) => isCategory(cat) && cat._id === categoryId);
      if (categoryIndex !== -1) {
        const category = newItems[categoryIndex] as Category;
        category.menuItems.push(item);
      }
      return { items: newItems };
    }),

  editMenuItem: (itemId, updatedItem) =>
    set((state) => {
      const newItems = [...state.items];
      const category = newItems.find(
        (cat) => isCategory(cat) && cat.menuItems.some((menuItem) => menuItem._id === itemId)
      ) as Category;
      if (category) {
        const itemIndex = category.menuItems.findIndex((menuItem) => menuItem._id === itemId);
        if (itemIndex !== -1) {
          category.menuItems[itemIndex] = updatedItem;
        }
      }
      return { items: newItems };
    }),

  deleteMenuItem: (categoryId, itemId) =>
    set((state) => {
      const newItems = [...state.items];
      const category = newItems.find(
        (cat) => isCategory(cat) && cat._id === categoryId
      ) as Category;
      if (category) {
        category.menuItems = category.menuItems.filter((item) => item._id !== itemId);
      }
      return { items: newItems };
    }),

  addCategory: (category) =>
    set((state) => ({
      items: [...state.items, category],
    })),

  editCategory: (categoryId, updatedCategory) =>
    set((state) => {
      const newItems = [...state.items];
      const categoryIndex = newItems.findIndex((cat) => isCategory(cat) && cat._id === categoryId);
      if (categoryIndex !== -1) {
        newItems[categoryIndex] = updatedCategory;
      }
      return { items: newItems };
    }),

  deleteCategory: (categoryId) =>
    set((state) => ({
      items: state.items.filter((item) => !(isCategory(item) && item._id === categoryId)),
    })),

  renameCategory: (categoryId, newName) =>
    set((state) => {
      const newItems = [...state.items];
      const category = newItems.find(
        (cat) => isCategory(cat) && cat._id === categoryId
      ) as Category;
      if (category) {
        category.name = newName;
      }
      return { items: newItems };
    }),

  addBanner: (banner) =>
    set((state) => ({
      items: [...state.items, banner],
    })),

  reset: () => set(initialState),
}));

export { isCategory };
export type { MenuItem, Category, Banner, Menu };
