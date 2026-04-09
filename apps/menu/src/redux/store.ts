import { combineReducers, configureStore } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from "redux-persist";
import storage from "redux-persist/lib/storage";
import configReducer from "./reducers/configSlice";
import menuReducer from "./reducers/menuSlice";
import themeReducer from "./reducers/themeSlice";

const persistConfig = {
  key: "rtylr:voxire:root",
  storage,
};

const rootReducer = combineReducers({
  config: configReducer,
  menu: menuReducer,
  theme: themeReducer,
});

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export const persistor = persistStore(store);
