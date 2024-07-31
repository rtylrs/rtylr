import { createSlice, createAsyncThunk } from "@reduxjs/toolkit";
import api from "../../lib/api";

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

export interface UserState {
  user: any | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: any[];
}

const initialState: UserState = {
  user: null,
  token: null,
  isAuthenticated: false,
  loading: false,
  error: null,
  role: null,
  userId: null,
  addons: {
    feedback: false,
    about: false,
    contact: false,
    landing: false,
  },
  visits: [],
};

export const loginAsync = createAsyncThunk(
  "user/loginAsync",
  async (
    payload: {
      email: string;
      password: string;
    },
    thunkAPI
  ) => {
    try {
      const response = await api.post("/auth/login", payload);
      const data = response.data;
      localStorage.setItem("token", data.token);

      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const registerRestaurantAsync = createAsyncThunk(
  "user/registerRestaurantAsync",
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
      const response = await api.post("/auth/create/business", payload);
      const data = response.data;
      localStorage.setItem("token", data.token);
      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const restaurantWithGoogleAsync = createAsyncThunk(
  "user/restaurantWithGoogleAsync",
  async (payload: { idToken: string }, thunkAPI) => {
    try {
      const response = await api.post("/auth/create/business/google", payload);
      const data = response.data;
      localStorage.setItem("token", data.token);
      return data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const getUserInfo = createAsyncThunk(
  "user/getUserInfo",
  async (_, thunkAPI) => {
    try {
      const response = await api.get(`/business/me`);
      return response.data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

export const setThemeAsync = createAsyncThunk(
  "user/setThemeAsync",
  async (
    payload: {
      id: string;
      themeData: {
        primaryColor?: string;
        secondaryColor?: string;
        loaderColor?: string;
        accentColor?: string;
        categoryTitleColor?: string;
        restaurantDescriptionColor?: string;
        productTitleColor?: string;
        productDescriptionColor?: string;
        productPriceColor?: string;
        menuBackgroundColor?: string;
        font?: string;
        fontSize?: string;
        categoryFontSize?: string;
        productPriceFontSize?: string;
        productDescriptionFontSize?: string;
        backgroundColor?: string;
        backgroundImage?: string;
        imageSize?: string;
        themeStyle?: string;
        displaySearchButton?: boolean;
      };
    },
    thunkAPI
  ) => {
    try {
      const response = await api.put(
        `/business/${payload.id}/theme`,
        payload.themeData
      );
      return response.data;
    } catch (error) {
      return thunkAPI.rejectWithValue((error as any).response.data);
    }
  }
);

const userSlice = createSlice({
  name: "user",
  initialState,
  reducers: {
    login: (state, action) => {
      state.user = action.payload;
      state.isAuthenticated = true;
    },
    logout: (state) => {
      state.user = null;
      state.isAuthenticated = false;
      localStorage.removeItem("token");
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
        const { token } = action.payload;
        state.token = token;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(loginAsync.rejected, (state, action) => {
        state.loading = false;
        state.error = (action.payload as { error: string }).error;
      })
      .addCase(registerRestaurantAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(registerRestaurantAsync.fulfilled, (state, action) => {
        const { token } = action.payload;
        state.token = token;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(registerRestaurantAsync.rejected, (state, action) => {
        state.loading = false;
        state.error = (action.payload as { error: string }).error;
      })
      .addCase(getUserInfo.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(getUserInfo.fulfilled, (state, action) => {
        const { userId, token, user, visits } = action.payload;
        return {
          ...state,
          role: user.role,
          userId: userId,
          token: token,
          user: user,
          isAuthenticated: true,
          loading: false,
          addons: user.addons,
          visits: visits,
        };
      })
      .addCase(getUserInfo.rejected, (state, action) => {
        state.loading = false;
        state.error = (action.payload as { error: string }).error;
      })
      .addCase(restaurantWithGoogleAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(restaurantWithGoogleAsync.fulfilled, (state, action) => {
        const { token } = action.payload;
        state.token = token;
        state.isAuthenticated = true;
        state.loading = false;
        state.error = null;
      })
      .addCase(restaurantWithGoogleAsync.rejected, (state, action) => {
        state.loading = false;
        state.error = (action.payload as { error: string }).error;
      })
      .addCase(setThemeAsync.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(setThemeAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.error = null;
        if (state.user) {
          state.user.theme = action.payload.theme;
        }
      })
      .addCase(setThemeAsync.rejected, (state, action) => {
        state.loading = false;
        state.error = (action.payload as { error: string }).error;
      });
  },
});

export const { login, logout } = userSlice.actions;

export const selectUser = (state: { user: UserState }) => state.user.user;
export const selectIsAuthenticated = (state: { user: UserState }) =>
  state.user.isAuthenticated;
export const selectLoading = (state: { user: UserState }) => state.user.loading;
export const selectError = (state: { user: UserState }) => state.user.error;
export const selectRole = (state: { user: UserState }) => state.user.role;
export const selectUserId = (state: { user: UserState }) => state.user.userId;
export const selectTheme = (state: { user: UserState }) =>
  state.user.user?.theme;

export default userSlice.reducer;
