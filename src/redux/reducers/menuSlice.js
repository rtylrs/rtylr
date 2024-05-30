import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import api from '../../lib/api';

export const fetchMenu = createAsyncThunk(
    'menu/fetchMenu',
    async (restaurantId) => {
        const response = await api.get(`/menu/${restaurantId}`);
        console.log(response.data);
        return response.data;
    }
);

export const addMenuItem = createAsyncThunk(
    'menu/addMenuItem',
    async ({ categoryId, item }) => {
        const response = await api.post(`/menu/category/${categoryId}/item`, item);
        return response.data;
    }
);

export const editMenuItem = createAsyncThunk(
    'menu/editMenuItem',
    async ({ itemId, item }) => {
        const response = await api.put(`/menu/item/${itemId}`, item);
        return response.data;
    }
);

export const deleteMenuItem = createAsyncThunk(
    'menu/deleteMenuItem',
    async ({ categoryId, itemId }) => {
        await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
        return { categoryId, itemId };
    }
);

export const addCategory = createAsyncThunk(
    'menu/addCategory',
    async (category) => {
        const response = await api.post('/menu/category', category);
        return response.data;
    }
);

export const deleteCategory = createAsyncThunk(
    'menu/deleteCategory',
    async (categoryId) => {
        await api.delete(`/menu/category/${categoryId}`);
        return categoryId;
    }
);

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
        status: 'idle',
        error: null,
    },
    reducers: {},
    extraReducers: (builder) => {
        builder
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
                const { categoryId, item } = action.payload;
                const category = state.items.find(cat => cat._id === categoryId && cat.type === 'category');
                if (category) {
                    const index = category.menuItems.findIndex(menuItem => menuItem._id === item._id);
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
