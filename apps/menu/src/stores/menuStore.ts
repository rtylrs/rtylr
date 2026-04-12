import { create } from 'zustand';
import api from '../shared/lib/api';

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

function isCategory(item: Category | Banner): item is Category {
  return (item as Category).type === 'category';
}

interface MenuStore {
  id: string | null;
  items: (Category | Banner)[];
  allMenus: Menu[];
  activeMenuId: string | null;
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;

  fetchAllMenus: (params: { restaurantId: string }) => Promise<void>;
  fetchMenu: (params: { userId: string; disabled: boolean }) => Promise<void>;
  fetchMenuById: (menuId: string) => Promise<void>;
  addMenuItem: (params: { categoryId: string; item: Record<string, any> }) => Promise<void>;
  editMenuItem: (params: { itemId: string; item: Partial<MenuItem> }) => Promise<void>;
  deleteMenuItem: (params: { categoryId: string; itemId: string }) => Promise<void>;
  addCategory: (category: Partial<Category>) => Promise<void>;
  editCategory: (params: { categoryId: string; category: Partial<Category> }) => Promise<void>;
  deleteCategory: (categoryId: string) => Promise<void>;
  renameCategory: (params: { categoryId: string; newName: string }) => Promise<void>;
  addBanner: (formData: FormData) => Promise<void>;
  setActiveMenuId: (params: { restaurantId: string; menuId: string }) => Promise<void>;
  createNewMenu: (params: { restaurantId: string; menuName: string }) => Promise<void>;
}

export const useMenuStore = create<MenuStore>()((set, get) => ({
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
  status: 'idle',
  error: null,

  fetchAllMenus: async ({ restaurantId }) => {
    set({ status: 'loading' });
    try {
      const response = await api.get(`/menu/menus/${restaurantId}`);
      if (response.data.message === 'No menus found') {
        set({ status: 'succeeded', allMenus: [] });
      } else {
        set({ status: 'succeeded', allMenus: response.data });
      }
    } catch (error) {
      set({
        status: 'failed',
        allMenus: [],
        error:
          (error as { response?: { data?: string } })?.response?.data ??
          (error as { message?: string })?.message ??
          null,
      });
    }
  },

  fetchMenu: async ({ userId, disabled }) => {
    set({ status: 'loading' });
    try {
      const response = await api.get(`/menu/${userId}`, {
        params: { disabled },
      });
      set({
        status: 'succeeded',
        items: response.data.items,
        id: response.data._id,
      });
    } catch (error) {
      set({
        status: 'failed',
        items: [],
        id: null,
        error: (error as { message?: string })?.message ?? null,
      });
    }
  },

  fetchMenuById: async (menuId) => {
    set({ status: 'loading' });
    try {
      const response = await api.get(`/menu/${menuId}`);
      set({
        status: 'succeeded',
        items: response.data.items,
        id: response.data._id,
      });
    } catch (error) {
      set({
        status: 'failed',
        items: [],
        id: null,
        error: (error as { message?: string })?.message ?? null,
      });
    }
  },

  addMenuItem: async ({ categoryId, item }) => {
    const formData = new FormData();
    Object.keys(item).forEach((key) => {
      if (Array.isArray(item[key])) {
        (item[key] as unknown[]).forEach((subItem, index: number) => {
          formData.append(`${key}[${index}]`, subItem as string | Blob);
        });
      } else {
        formData.append(key, item[key]);
      }
    });

    const response = await api.post(`/menu/category/${categoryId}/item`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });

    set((state) => {
      const items = state.items.map((cat) => {
        if (isCategory(cat) && cat._id === categoryId) {
          return { ...cat, menuItems: [...cat.menuItems, response.data] };
        }
        return cat;
      });
      return { items };
    });
  },

  editMenuItem: async ({ itemId, item }) => {
    const response = await api.put(`/menu/item/${itemId}`, item);

    set((state) => {
      const items = state.items.map((cat) => {
        if (isCategory(cat) && cat.menuItems.some((mi) => mi._id === itemId)) {
          return {
            ...cat,
            menuItems: cat.menuItems.map((mi) => (mi._id === itemId ? response.data : mi)),
          };
        }
        return cat;
      });
      return { items };
    });
  },

  deleteMenuItem: async ({ categoryId, itemId }) => {
    await api.delete(`/menu/category/${categoryId}/item/${itemId}`);

    set((state) => {
      const items = state.items.map((cat) => {
        if (isCategory(cat) && cat._id === categoryId) {
          return {
            ...cat,
            menuItems: cat.menuItems.filter((mi) => mi._id !== itemId),
          };
        }
        return cat;
      });
      return { items };
    });
  },

  addCategory: async (category) => {
    const response = await api.post('/menu/category', category);
    set((state) => ({ items: [...state.items, response.data] }));
  },

  editCategory: async ({ categoryId, category }) => {
    const response = await api.put(`/menu/category/${categoryId}`, category);
    set((state) => {
      const items = state.items.map((cat) => {
        if (isCategory(cat) && cat._id === categoryId) {
          return response.data;
        }
        return cat;
      });
      return { items };
    });
  },

  deleteCategory: async (categoryId) => {
    await api.delete(`/menu/category/${categoryId}`);
    set((state) => ({
      items: state.items.filter((item) => isCategory(item) && item._id !== categoryId),
    }));
  },

  renameCategory: async ({ categoryId, newName }) => {
    const response = await api.put(`/menu/category/${categoryId}`, { newName });
    set((state) => {
      const items = state.items.map((cat) => {
        if (isCategory(cat) && cat._id === response.data._id) {
          return { ...cat, name: response.data.name };
        }
        return cat;
      });
      return { items };
    });
  },

  addBanner: async (formData) => {
    const response = await api.post('/menu/banner', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    set((state) => ({ items: [...state.items, response.data] }));
  },

  setActiveMenuId: async ({ restaurantId, menuId }) => {
    await api.post('/menu/set-active-menu', { restaurantId, menuId });
    set({ activeMenuId: menuId, id: menuId, items: [] });
  },

  createNewMenu: async ({ restaurantId, menuName }) => {
    const response = await api.post('/menu/createNewMenu', {
      restaurantId,
      menuName,
    });
    set((state) => ({ allMenus: [...state.allMenus, response.data] }));
  },
}));
