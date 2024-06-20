import { combineReducers, configureStore } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from "redux-persist";
import storage from "redux-persist/lib/storage";
import userReducer, { UserState } from "./reducers/userSlice";
import cartReducer, { CartState } from "./reducers/cartSlice";
import configReducer, { ConfigState } from "./reducers/configSlice";
import menuReducer, { MenuState } from "./reducers/menuSlice";
import { thunk } from "redux-thunk";

export interface RootState {
  user: UserState;
  cart: CartState;
  config: ConfigState;
  menu: MenuState;
}

const persistConfig = {
  key: "vbitez:voxire:root",
  storage,
};

const reducers = combineReducers({
  user: userReducer,
  cart: cartReducer,
  config: configReducer,
  menu: menuReducer,
});

const thunkMiddleware: any = thunk;

export const store = configureStore({
  reducer: persistReducer(persistConfig, reducers),
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
      thunk: thunkMiddleware,
    }),
});

export const persistor = persistStore(store);
