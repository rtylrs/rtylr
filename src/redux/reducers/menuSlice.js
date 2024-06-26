import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import api from '../../lib/api';


// Fetch all menus by restaurant ID
export const fetchAllMenus = createAsyncThunk(
    'menu/fetchAllMenus',
    async ({ restaurantId }) => {
        const response = await api.get(`/menu/menus/${restaurantId}`);
        console.log(response.data);
        return response.data;
    }
);

// Set active menu ID
export const setActiveMenuId = createAsyncThunk(
    'menu/setActiveMenuId',
    async ({ restaurantId, menuId }) => {
        const response = await api.post('/menu/set-active-menu', { restaurantId, menuId });
        return { restaurantId, menuId };
    }
);

export const createNewMenu = createAsyncThunk(
    'menu/createNewMenu',
    async ({ restaurantId, menuName }) => {
        console.log(restaurantId, menuName);
        const response = await api.post('/menu/createNewMenu', { restaurantId, menuName });
        return response.data;
    }
);

// Fetch menu by menu ID
export const fetchMenuById = createAsyncThunk(
    'menu/fetchMenuById',
    async (menuId) => {
        const response = await api.get(`/menu/${menuId}`);
        return response.data;
    }
);

// Fetch menu by restaurant ID
export const fetchMenu = createAsyncThunk(
    'menu/fetchMenu',
    async (restaurantId) => {
        const response = await api.get(`/menu/${restaurantId}`);
        return response.data;
    }
);

// Add a new menu item to a category
export const addMenuItem = createAsyncThunk(
    'menu/addMenuItem',
    async ({ categoryId, item }) => {
        const formData = new FormData();
        Object.keys(item).forEach(key => {
            if (Array.isArray(item[key])) {
                item[key].forEach((subItem, index) => {
                    formData.append(`${key}[${index}]`, subItem);
                });
            } else {
                formData.append(key, item[key]);
            }
        });

        const response = await api.post(`/menu/category/${categoryId}/item`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
        return { categoryId, item: response.data };
    }
);

// Add a new banner
export const addBanner = createAsyncThunk(
    'menu/addBanner',
    async (formData) => {
        const response = await api.post('/menu/banner', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
        return response.data;
    }
);

// Edit a menu item
export const editMenuItem = createAsyncThunk(
    'menu/editMenuItem',
    async ({ itemId, item }) => {
        const response = await api.put(`/menu/item/${itemId}`, item);
        return { itemId, item: response.data };
    }
);

// Delete a menu item from a category
export const deleteMenuItem = createAsyncThunk(
    'menu/deleteMenuItem',
    async ({ categoryId, itemId }) => {
        await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
        return { categoryId, itemId };
    }
);

// Add a new category
export const addCategory = createAsyncThunk(
    'menu/addCategory',
    async (category) => {
        const response = await api.post('/menu/category', category);
        return response.data;
    }
);

// Delete a category
export const deleteCategory = createAsyncThunk(
    'menu/deleteCategory',
    async (categoryId) => {
        await api.delete(`/menu/category/${categoryId}`);
        return categoryId;
    }
);

// Rename a category
export const renameCategory = createAsyncThunk(
    'menu/renameCategory',
    async ({ categoryId, newName }) => {
        const response = await api.put(`/menu/category/${categoryId}`, { newName });
        return response.data;
    }
);

const menuSlice = createSlice({
    name: 'menu',
    initialState: {
        id: null,
        items: [],
        allMenus: [],
        activeMenuId: null,
        status: 'idle',
        error: null,
    },
    reducers: {},
    extraReducers: (builder) => {
        builder
            .addCase(fetchAllMenus.pending, (state) => {
                state.status = 'loading';
            })
            .addCase(fetchAllMenus.fulfilled, (state, action) => {
                state.status = 'succeeded';
                state.allMenus = action.payload;
            })
            .addCase(fetchAllMenus.rejected, (state, action) => {
                state.status = 'failed';
                state.allMenus = [];
                state.error = action.error.message;
            })
            .addCase(setActiveMenuId.fulfilled, (state, action) => {
                state.activeMenuId = action.payload.menuId;
                state.id = action.payload.menuId;
                state.items = [];
            })
            .addCase(createNewMenu.fulfilled, (state, action) => {
                state.allMenus.push(action.payload);
            })
            .addCase(fetchMenuById.pending, (state) => {
                state.status = 'loading';
            })
            .addCase(fetchMenuById.fulfilled, (state, action) => {
                state.status = 'succeeded';
                state.items = action.payload.items;
                state.id = action.payload._id;
            })
            .addCase(fetchMenuById.rejected, (state, action) => {
                state.status = 'failed';
                state.items = [];
                state.id = null;
                state.error = action.error.message;
            })
            .addCase(fetchMenu.pending, (state) => {
                state.status = 'loading';
            })
            .addCase(fetchMenu.fulfilled, (state, action) => {
                state.status = 'succeeded';
                state.items = action.payload.items;
                state.id = action.payload._id;
            })
            .addCase(fetchMenu.rejected, (state, action) => {
                state.status = 'failed';
                state.items = [];
                state.id = null;
                state.error = action.error.message;
            })
            .addCase(addMenuItem.fulfilled, (state, action) => {
                const { categoryId, item } = action.payload;
                const category = state.items.find(cat => cat._id === categoryId && cat.type === 'category');
                if (category) {
                    category.menuItems.push(item);
                }
            })
            .addCase(editMenuItem.fulfilled, (state, action) => {
                const { itemId, item } = action.payload;
                const category = state.items.find(cat => cat.menuItems.some(menuItem => menuItem._id === itemId) && cat.type === 'category');
                if (category) {
                    const index = category.menuItems.findIndex(menuItem => menuItem._id === itemId);
                    if (index !== -1) {
                        category.menuItems[index] = item;
                    }
                }
            })
            .addCase(deleteMenuItem.fulfilled, (state, action) => {
                const { categoryId, itemId } = action.payload;
                const category = state.items.find(cat => cat._id === categoryId && cat.type === 'category');
                if (category) {
                    category.menuItems = category.menuItems.filter(item => item._id !== itemId);
                }
            })
            .addCase(addCategory.fulfilled, (state, action) => {
                state.items.push(action.payload);
            })
            .addCase(deleteCategory.fulfilled, (state, action) => {
                state.items = state.items.filter(item => item._id !== action.payload && item.type === 'category');
            })
            .addCase(renameCategory.fulfilled, (state, action) => {
                const { _id, name } = action.payload;
                const category = state.items.find(cat => cat._id === _id && cat.type === 'category');
                if (category) {
                    category.name = name;
                }
            });
    }
});


export default menuSlice.reducer;
