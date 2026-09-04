import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { tenantStorage } from '../utils/tenantStorage';

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

interface CategoryDetails {
  menuItems?: MenuItem[];
  [key: string]: unknown;
}

interface Category {
  _id: string;
  /** Legacy wire id alias; canonical identity is `_id`. */
  id?: string;
  /** Legacy wire id alias; canonical identity is `_id`. */
  itemId?: string;
  name: string;
  type: 'category';
  menuItems: MenuItem[];
  details?: CategoryDetails;
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
  deleteBanner: (id: string) => void;
  addMenuItem: (categoryId: string, item: Partial<MenuItem>) => Promise<void>;
  updateMenuItem: (id: string, updates: Partial<MenuItem>) => void;
  deleteMenuItem: (id: string) => void;
}

function isCategory(item: Category | Banner | MenuItem): item is Category {
  return 'type' in item && item.type === 'category';
}

function isMenuItem(item: Category | Banner | MenuItem): item is MenuItem {
  return !isCategory(item) && !('bannerImage' in item);
}

/** Category identity covers the canonical `_id` plus legacy wire aliases. */
function matchesCategoryId(item: Category, id: string): boolean {
  return item._id === id || item.id === id || item.itemId === id;
}

/** MenuItem identity covers the canonical `_id` plus the legacy alias. */
function matchesMenuItemId(item: MenuItem, id: string): boolean {
  return item._id === id || item.itemId === id;
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
            isCategory(item) && matchesCategoryId(item, id) ? { ...item, ...updates } : item
          ),
        })),

      addBanner: (banner) => set((state) => ({ items: [...state.items, banner] })),

      deleteCategory: (id) =>
        set((state) => ({
          items: state.items.filter((item) => !isCategory(item) || !matchesCategoryId(item, id)),
        })),

      deleteBanner: (id) =>
        set((state) => ({
          items: state.items.filter((item) => item._id !== id),
        })),

      addMenuItem: async (categoryId, item) => {
        const name = typeof item.name === 'string' && item.name ? item.name : 'New Product';
        const description = typeof item.description === 'string' ? item.description : '';
        const rawPrice: unknown = item.price;
        const id = String(Date.now());
        const newItem: MenuItem = {
          _id: id,
          itemId: id,
          type: 'item',
          name,
          details: {
            name,
            description,
            price: typeof rawPrice === 'number' ? rawPrice : 0,
            isActive: item.isAvailable !== false,
          },
          ...item,
        };
        set((state) => {
          const idx = state.items.findIndex(
            (entry) => isCategory(entry) && matchesCategoryId(entry, categoryId)
          );
          if (idx === -1) return { items: [...state.items, newItem] };
          const current = state.items[idx];
          if (!isCategory(current)) return { items: [...state.items, newItem] };
          const updated: Category = {
            ...current,
            details: {
              ...(current.details ?? {}),
              menuItems: [...(current.details?.menuItems ?? []), newItem],
            },
            menuItems: [...current.menuItems, newItem],
          };
          const items = [...state.items];
          items[idx] = updated;
          return { items };
        });
      },

      updateMenuItem: (id, updates) =>
        set((state) => ({
          items: state.items.map((item): Category | Banner | MenuItem => {
            // The top-level pass is menu items only. It used to match any
            // entry in the union by `_id`, cast the result to MenuItem and
            // overwrite `details` with `{ name }`, which for a category
            // whose id collided wiped `details.menuItems` outright and for a
            // banner produced an entry that was neither shape. Both were
            // reachable only through an id collision, and the `as any` casts
            // are what allowed them to compile.
            if (isMenuItem(item) && matchesMenuItemId(item, id)) {
              const currentDetails = item.details ?? { name: item.name || '' };
              return { ...item, ...updates, details: { ...currentDetails, ...updates.details } };
            }
            if (isCategory(item) && item.details?.menuItems) {
              const updatedSubItems = item.details.menuItems.map((subItem) => {
                if (!matchesMenuItemId(subItem, id)) return subItem;
                const currentSubDetails = subItem.details ?? { name: subItem.name || '' };
                return {
                  ...subItem,
                  ...updates,
                  details: { ...currentSubDetails, ...updates.details },
                };
              });
              return {
                ...item,
                details: { ...item.details, menuItems: updatedSubItems },
              };
            }
            return item;
          }),
        })),

      deleteMenuItem: (id) =>
        set((state) => ({
          items: state.items
            .filter((item) => item._id !== id && (!isMenuItem(item) || item.itemId !== id))
            .map((item) => {
              if (!isCategory(item)) return item;
              const prunedLegacy = item.menuItems.filter((sub) => !matchesMenuItemId(sub, id));
              const nested = item.details?.menuItems;
              const prunedNested = nested?.filter((sub) => !matchesMenuItemId(sub, id));
              if (!prunedNested && prunedLegacy.length === item.menuItems.length) return item;
              const next: Category = { ...item, menuItems: prunedLegacy };
              if (prunedNested) {
                next.details = { ...(item.details ?? {}), menuItems: prunedNested };
              }
              return next;
            }),
        })),
    }),
    {
      name: 'menu-store',
      storage: tenantStorage,
      partialize: (state) => ({
        id: state.id,
        items: state.items,
        allMenus: state.allMenus,
        activeMenuId: state.activeMenuId,
      }),
    }
  )
);
