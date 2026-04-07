import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../lib/api';

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
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;
}

interface MenuActions {
  setId: (id: string | null) => void;
  setItems: (items: (Category | Banner)[]) => void;
  setAllMenus: (menus: Menu[]) => void;
  setActiveMenuId: (id: string | null) => void;
  setStatus: (status: 'idle' | 'loading' | 'succeeded' | 'failed') => void;
  setError: (error: string | null) => void;
  fetchAllMenus: (restaurantId: string) => Promise<void>;
  setActiveMenu: (restaurantId: string, menuId: string) => Promise<void>;
  createNewMenu: (restaurantId: string, menuName: string) => Promise<void>;
  fetchMenuById: (menuId: string) => Promise<void>;
  fetchMenu: (userId: string, disabled: boolean) => Promise<void>;
  addMenuItem: (categoryId: string, item: Record<string, any>) => Promise<void>;
  addBanner: (formData: FormData) => Promise<void>;
  editMenuItem: (itemId: string, item: Partial<MenuItem>) => Promise<void>;
  deleteMenuItem: (categoryId: string, itemId: string) => Promise<void>;
  addCategory: (category: Partial<Category>) => Promise<void>;
  editCategory: (categoryId: string, category: Partial<Category>) => Promise<void>;
  deleteCategory: (categoryId: string) => Promise<void>;
  renameCategory: (categoryId: string, newName: string) => Promise<void>;
}

function isCategory(item: Category | Banner): item is Category {
  return (item as Category).type === 'category';
}

const initialState: MenuState = {
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
  status: 'idle',
  error: null,
};

export const useMenuStore = create<MenuState & MenuActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setId: (id) => set({ id }),

      setItems: (items) => set({ items }),

      setAllMenus: (allMenus) => set({ allMenus }),

      setActiveMenuId: (activeMenuId) => set({ activeMenuId }),

      setStatus: (status) => set({ status }),

      setError: (error) => set({ error }),

      fetchAllMenus: async (restaurantId) => {
        set({ status: 'loading' });
        try {
          const response = await api.get(`/menu/menus/${restaurantId}`);
          if (response.data.message === 'No menus found') {
            set({ allMenus: [], status: 'succeeded' });
            return;
          }
          set({ allMenus: response.data, status: 'succeeded' });
        } catch (error: any) {
          set({
            status: 'failed',
            allMenus: [],
            error: error.response?.data || error.message
          });
        }
      },

      setActiveMenu: async (restaurantId, menuId) => {
        try {
          await api.post('/menu/set-active-menu', { restaurantId, menuId });
          set({ activeMenuId: menuId, id: menuId, items: [] });
        } catch (error) {
          console.error('Error setting active menu:', error);
        }
      },

      createNewMenu: async (restaurantId, menuName) => {
        try {
          const response = await api.post('/menu/createNewMenu', {
            restaurantId,
            menuName,
          });
          const state = get();
          set({ allMenus: [...state.allMenus, response.data] });
        } catch (error) {
          console.error('Error creating new menu:', error);
        }
      },

      fetchMenuById: async (menuId) => {
        set({ status: 'loading' });
        try {
          const response = await api.get(`/menu/${menuId}`);
          set({
            status: 'succeeded',
            items: response.data.items,
            id: response.data._id
          });
        } catch (error: any) {
          set({
            status: 'failed',
            items: [],
            id: null,
            error: error.message
          });
        }
      },

      fetchMenu: async (userId, disabled) => {
        set({ status: 'loading' });
        try {
          const response = await api.get(`/menu/${userId}`, {
            params: { disabled },
          });
          set({
            status: 'succeeded',
            items: response.data.items,
            id: response.data._id
          });
        } catch (error: any) {
          set({
            status: 'failed',
            items: [],
            id: null,
            error: error.message
          });
        }
      },

      addMenuItem: async (categoryId, item) => {
        try {
          const formData = new FormData();
          Object.keys(item).forEach((key) => {
            if (Array.isArray(item[key])) {
              item[key].forEach((subItem: any, index: number) => {
                formData.append(`${key}[${index}]`, subItem);
              });
            } else {
              formData.append(key, item[key]);
            }
          });

          const response = await api.post(
            `/menu/category/${categoryId}/item`,
            formData,
            {
              headers: {
                'Content-Type': 'multipart/form-data',
              },
            }
          );

          const state = get();
          const updatedItems = state.items.map((cat) => {
            if (isCategory(cat) && cat._id === categoryId) {
              return {
                ...cat,
                menuItems: [...cat.menuItems, response.data],
              };
            }
            return cat;
          });
          set({ items: updatedItems });
        } catch (error) {
          console.error('Error adding menu item:', error);
        }
      },

      addBanner: async (formData) => {
        try {
          const response = await api.post('/menu/banner', formData, {
            headers: {
              'Content-Type': 'multipart/form-data',
            },
          });
          const state = get();
          set({ items: [...state.items, response.data] });
        } catch (error) {
          console.error('Error adding banner:', error);
        }
      },

      editMenuItem: async (itemId, item) => {
        try {
          const response = await api.put(`/menu/item/${itemId}`, item);
          const state = get();
          const updatedItems = state.items.map((cat) => {
            if (isCategory(cat)) {
              const updatedMenuItems = cat.menuItems.map((menuItem) =>
                menuItem._id === itemId ? response.data : menuItem
              );
              return { ...cat, menuItems: updatedMenuItems };
            }
            return cat;
          });
          set({ items: updatedItems });
        } catch (error) {
          console.error('Error editing menu item:', error);
        }
      },

      deleteMenuItem: async (categoryId, itemId) => {
        try {
          await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
          const state = get();
          const updatedItems = state.items.map((cat) => {
            if (isCategory(cat) && cat._id === categoryId) {
              return {
                ...cat,
                menuItems: cat.menuItems.filter((item) => item._id !== itemId),
              };
            }
            return cat;
          });
          set({ items: updatedItems });
        } catch (error) {
          console.error('Error deleting menu item:', error);
        }
      },

      addCategory: async (category) => {
        try {
          const response = await api.post('/menu/category', category);
          const state = get();
          set({ items: [...state.items, response.data] });
        } catch (error) {
          console.error('Error adding category:', error);
        }
      },

      editCategory: async (categoryId, category) => {
        try {
          const response = await api.put(`/menu/category/${categoryId}`, category);
          const state = get();
          const updatedItems = state.items.map((item) =>
            isCategory(item) && item._id === categoryId ? response.data : item
          );
          set({ items: updatedItems });
        } catch (error) {
          console.error('Error editing category:', error);
        }
      },

      deleteCategory: async (categoryId) => {
        try {
          await api.delete(`/menu/category/${categoryId}`);
          const state = get();
          const updatedItems = state.items.filter(
            (item) => !(isCategory(item) && item._id === categoryId)
          );
          set({ items: updatedItems });
        } catch (error: any) {
          console.error('Error deleting category:', error);
          throw error;
        }
      },

      renameCategory: async (categoryId, newName) => {
        try {
          const response = await api.put(`/menu/category/${categoryId}`, { newName });
          const state = get();
          const updatedItems = state.items.map((cat) => {
            if (isCategory(cat) && cat._id === categoryId) {
              return { ...cat, name: response.data.name };
            }
            return cat;
          });
          set({ items: updatedItems });
        } catch (error) {
          console.error('Error renaming category:', error);
        }
      },
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