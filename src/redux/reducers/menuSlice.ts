import { createSlice, createAsyncThunk, PayloadAction } from "@reduxjs/toolkit";
import api from "../../lib/api";

interface MenuItem {
  _id: string;
  name: string;
  [key: string]: any;
}

interface Category {
  _id: string;
  name: string;
  type: "category";
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
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
}

function isCategory(item: Category | Banner): item is Category {
  return (item as Category).type === "category";
}

// Define initial state
const initialState: MenuState = {
  id: null,
  items: [],
  allMenus: [],
  activeMenuId: null,
  status: "idle",
  error: null,
};

export const fetchAllMenus = createAsyncThunk<Menu[], { restaurantId: string }>(
  "menu/fetchAllMenus",
  async ({ restaurantId }) => {
    const response = await api.get(`/menu/menus/${restaurantId}`);
    return response.data;
  }
);

export const setActiveMenuId = createAsyncThunk<
  { restaurantId: string; menuId: string },
  { restaurantId: string; menuId: string }
>("menu/setActiveMenuId", async ({ restaurantId, menuId }) => {
  await api.post("/menu/set-active-menu", { restaurantId, menuId });
  return { restaurantId, menuId };
});

export const createNewMenu = createAsyncThunk<
  Menu,
  { restaurantId: string; menuName: string }
>("menu/createNewMenu", async ({ restaurantId, menuName }) => {
  const response = await api.post("/menu/createNewMenu", {
    restaurantId,
    menuName,
  });
  return response.data;
});

export const fetchMenuById = createAsyncThunk<Menu, string>(
  "menu/fetchMenuById",
  async (menuId) => {
    const response = await api.get(`/menu/${menuId}`);
    return response.data;
  }
);

export const fetchMenu = createAsyncThunk<Menu, string>(
  "menu/fetchMenu",
  async (restaurantId) => {
    const response = await api.get(`/menu/${restaurantId}`);
    return response.data;
  }
);

export const addMenuItem = createAsyncThunk<
  { categoryId: string; item: MenuItem },
  { categoryId: string; item: Record<string, any> }
>("menu/addMenuItem", async ({ categoryId, item }) => {
  const formData = new FormData();
  Object.keys(item).forEach((key) => {
    if (Array.isArray(item[key])) {
      item[key].forEach((subItem, index) => {
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
        "Content-Type": "multipart/form-data",
      },
    }
  );
  return { categoryId, item: response.data };
});

export const addBanner = createAsyncThunk<Banner, FormData>(
  "menu/addBanner",
  async (formData) => {
    const response = await api.post("/menu/banner", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return response.data;
  }
);

export const editMenuItem = createAsyncThunk<
  { itemId: string; item: MenuItem },
  { itemId: string; item: Partial<MenuItem> }
>("menu/editMenuItem", async ({ itemId, item }) => {
  const response = await api.put(`/menu/item/${itemId}`, item);
  return { itemId, item: response.data };
});

export const deleteMenuItem = createAsyncThunk<
  { categoryId: string; itemId: string },
  { categoryId: string; itemId: string }
>("menu/deleteMenuItem", async ({ categoryId, itemId }) => {
  await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
  return { categoryId, itemId };
});

export const addCategory = createAsyncThunk<Category, Partial<Category>>(
  "menu/addCategory",
  async (category) => {
    const response = await api.post("/menu/category", category);
    return response.data;
  }
);

export const editCategory = createAsyncThunk<
  { categoryId: string; category: Category },
  { categoryId: string; category: Partial<Category> }
>("menu/editCategory", async ({ categoryId, category }) => {
  const response = await api.put(`/menu/category/${categoryId}`, category);
  return { categoryId, category: response.data };
});

export const deleteCategory = createAsyncThunk<string, string>(
  "menu/deleteCategory",
  async (categoryId) => {
    await api.delete(`/menu/category/${categoryId}`);
    return categoryId;
  }
);

export const renameCategory = createAsyncThunk<
  Category,
  { categoryId: string; newName: string }
>("menu/renameCategory", async ({ categoryId, newName }) => {
  const response = await api.put(`/menu/category/${categoryId}`, { newName });
  return response.data;
});

const menuSlice = createSlice({
  name: "menu",
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchAllMenus.pending, (state) => {
        state.status = "loading";
      })
      .addCase(
        fetchAllMenus.fulfilled,
        (state, action: PayloadAction<Menu[]>) => {
          state.status = "succeeded";
          state.allMenus = action.payload;
        }
      )
      .addCase(fetchAllMenus.rejected, (state, action) => {
        state.status = "failed";
        state.allMenus = [];
        state.error = action.error.message || null;
      })
      .addCase(
        setActiveMenuId.fulfilled,
        (
          state,
          action: PayloadAction<{ restaurantId: string; menuId: string }>
        ) => {
          state.activeMenuId = action.payload.menuId;
          state.id = action.payload.menuId;
          state.items = [];
        }
      )
      .addCase(
        createNewMenu.fulfilled,
        (state, action: PayloadAction<Menu>) => {
          state.allMenus.push(action.payload);
        }
      )
      .addCase(fetchMenuById.pending, (state) => {
        state.status = "loading";
      })
      .addCase(
        fetchMenuById.fulfilled,
        (state, action: PayloadAction<Menu>) => {
          state.status = "succeeded";
          state.items = action.payload.items;
          state.id = action.payload._id;
        }
      )
      .addCase(fetchMenuById.rejected, (state, action) => {
        state.status = "failed";
        state.items = [];
        state.id = null;
        state.error = action.error.message || null;
      })
      .addCase(fetchMenu.pending, (state) => {
        state.status = "loading";
      })
      .addCase(fetchMenu.fulfilled, (state, action: PayloadAction<Menu>) => {
        state.status = "succeeded";
        state.items = action.payload.items;
        state.id = action.payload._id;
      })
      .addCase(fetchMenu.rejected, (state, action) => {
        state.status = "failed";
        state.items = [];
        state.id = null;
        state.error = action.error.message || null;
      })
      .addCase(
        addMenuItem.fulfilled,
        (
          state,
          action: PayloadAction<{ categoryId: string; item: MenuItem }>
        ) => {
          const { categoryId, item } = action.payload;
          const category = state.items.find(
            (cat) => isCategory(cat) && cat._id === categoryId
          ) as Category;
          if (category) {
            category.menuItems.push(item);
          }
        }
      )
      .addCase(
        editMenuItem.fulfilled,
        (state, action: PayloadAction<{ itemId: string; item: MenuItem }>) => {
          const { itemId, item } = action.payload;
          const category = state.items.find(
            (cat) =>
              isCategory(cat) &&
              cat.menuItems.some((menuItem) => menuItem._id === itemId)
          ) as Category;
          if (category) {
            const index = category.menuItems.findIndex(
              (menuItem) => menuItem._id === itemId
            );
            if (index !== -1) {
              category.menuItems[index] = item;
            }
          }
        }
      )
      .addCase(
        deleteMenuItem.fulfilled,
        (
          state,
          action: PayloadAction<{ categoryId: string; itemId: string }>
        ) => {
          const { categoryId, itemId } = action.payload;
          const category = state.items.find(
            (cat) => isCategory(cat) && cat._id === categoryId
          ) as Category;
          if (category) {
            category.menuItems = category.menuItems.filter(
              (item) => item._id !== itemId
            );
          }
        }
      )
      .addCase(
        addCategory.fulfilled,
        (state, action: PayloadAction<Category>) => {
          state.items.push(action.payload);
        }
      )
      .addCase(
        deleteCategory.fulfilled,
        (state, action: PayloadAction<string>) => {
          state.items = state.items.filter(
            (item) => isCategory(item) && item._id !== action.payload
          );
        }
      )
      .addCase(
        renameCategory.fulfilled,
        (state, action: PayloadAction<Category>) => {
          const { _id, name } = action.payload;
          const category = state.items.find(
            (cat) => isCategory(cat) && cat._id === _id
          ) as Category;
          if (category) {
            category.name = name;
          }
        }
      );
  },
});

export default menuSlice.reducer;
