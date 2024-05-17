import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import api from '../../lib/api';

const initialState = {
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false,
    error: null,
    role: null,
    userId: null,
};

// Async thunk for logging in
export const loginAsync = createAsyncThunk(
    'user/loginAsync',
    async ({ email, password }, thunkAPI) => {
        try {
            const response = await api.post('/auth/login', { email, password });
            const data = response.data;
            localStorage.setItem('token', data.token);
            return data;
        } catch (error) {
            return thunkAPI.rejectWithValue(error.response.data);
        }
    }
);

export const RegisterResaurantAsync = createAsyncThunk(
    'user/RegisterResaurantAsync',
    async ({ email, password, name, refered }, thunkAPI) => {
        try {
            const response = await api.post('/auth/create/business', {
                email,
                password,
                name,
                refered: refered.length !== 10 ? null : refered
            });
            const data = response.data;
            localStorage.setItem('token', data.token);
            return data;
        } catch (error) {
            return thunkAPI.rejectWithValue(error.response.data);
        }
    }
);

const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {
        login: (state, action) => {
            state.user = action.payload;
            state.isAuthenticated = true;
        },
        logout: (state) => {
            state.user = null;
            state.isAuthenticated = false;
            localStorage.removeItem('token');
            state.loading = false;
            state.error = null;
            state.role = null;
            state.userId = null;
            state.token = null;
        },
    },
    extraReducers: (builder) => {
        builder
            .addCase(loginAsync.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(loginAsync.fulfilled, (state, action) => {
                const { userId, token, user } = action.payload;
                return {
                    ...state,
                    role: user.role,
                    userId: userId,
                    token: token,
                    user: user,
                    isAuthenticated: true,
                    loading: false,
                };
            })
            .addCase(loginAsync.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload;
            }).addCase(RegisterResaurantAsync.pending, (state) => {
                state.loading = true;
                state.error = null;
            }).addCase(RegisterResaurantAsync.fulfilled, (state, action) => {
                const { userId, token, user } = action.payload;
                return {
                    ...state,
                    role: user.role,
                    userId: userId,
                    token: token,
                    user: user,
                    isAuthenticated: true,
                    loading: false,
                };
            }).addCase(RegisterResaurantAsync.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload;
            });
    },
});

export const { login, logout } = userSlice.actions;

export const selectUser = (state) => state.user.user;
export const selectIsAuthenticated = (state) => state.user.isAuthenticated;
export const selectLoading = (state) => state.user.loading;
export const selectError = (state) => state.user.error;
export const selectRole = (state) => state.user.role;
export const selectUserId = (state) => state.user.userId;

export default userSlice.reducer;
