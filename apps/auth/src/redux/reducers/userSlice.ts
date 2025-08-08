import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import api from '../../lib/api';
import {
  setCookie,
  removeCookie,
  getCookie,
  getDefaultCookieDomain,
} from '../../lib/cookies';

export interface UserState {
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  role: string | null;
}

const initialState: UserState = {
  token: null,
  isAuthenticated: false,
  error: null,
  loading: false,
  role: null,
};

export const loginAsync = createAsyncThunk(
  'user/loginAsync',
  async (
    payload: {
      email: string;
      password: string;
    },
    thunkAPI
  ) => {
    try {
      const response = await api.post('/auth/login', payload);
      const data = response.data;
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });

      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const registerRestaurantAsync = createAsyncThunk(
  'user/registerRestaurantAsync',
  async (
    payload: {
      email: string;
      password: string;
      name: string;
      referred: string;
    },
    thunkAPI
  ) => {
    try {
      const response = await api.post('/auth/create/business', payload);
      const data = response.data;
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });
      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const restaurantWithGoogleAsync = createAsyncThunk(
  'user/restaurantWithGoogleAsync',
  async (payload: { idToken: string }, thunkAPI) => {
    try {
      const response = await api.post('/auth/create/business/google', payload);
      const data = response.data;
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });
      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    login: (state, action) => {
      // Token is stored in cookie only (not persisted in Redux)
      state.token = null;
      state.role = action.payload.role;
      state.isAuthenticated = true;
    },
    logout: (state) => {
      state.isAuthenticated = false;
      removeCookie('token', { domain: getDefaultCookieDomain() });
      state.loading = false;
      state.error = null;
      state.role = null;
      state.token = null;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(loginAsync.fulfilled, (state, action) => {
        const { role } = action.payload;
        state.token = null;
        state.role = role;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(loginAsync.rejected, (state, action) => {
        state.loading = false;
        const payload = action.payload as any;
        state.error = payload?.message || payload?.error || 'Login failed';
      })
      .addCase(registerRestaurantAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(registerRestaurantAsync.fulfilled, (state, action) => {
        const { role } = action.payload;
        state.token = null;
        state.role = role;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(registerRestaurantAsync.rejected, (state, action) => {
        state.loading = false;
        const payload = action.payload as any;
        state.error =
          payload?.message || payload?.error || 'Registration failed';
      })
      .addCase(restaurantWithGoogleAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(restaurantWithGoogleAsync.fulfilled, (state, action) => {
        const { role } = action.payload;
        state.token = null;
        state.role = role;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(restaurantWithGoogleAsync.rejected, (state, action) => {
        state.loading = false;
        const payload = action.payload as any;
        state.error =
          payload?.message || payload?.error || 'Google login failed';
      });
  },
});

export const { login, logout, clearError } = userSlice.actions;

export const selectIsAuthenticated = (state: { user: UserState }) =>
  state.user.isAuthenticated;
export const selectLoading = (state: { user: UserState }) => state.user.loading;
export const selectError = (state: { user: UserState }) => state.user.error;
export const selectRole = (state: { user: UserState }) => state.user.role;
export const selectToken = (_state: { user: UserState }) => getCookie('token');

export default userSlice.reducer;
