import { combineReducers, configureStore, getDefaultMiddleware } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import userReducer from './reducers/userSlice';
import productsReducer from './reducers/productsSlice';
import cartReducer from './reducers/cartSlice';
import configReducer from './reducers/configSlice';
import { thunk } from 'redux-thunk';

const persistConfig = {
    key: 'vbitez:voxire:root',
    storage,
};

const reducers = combineReducers({
    user: userReducer,
    products: productsReducer,
    cart: cartReducer,
    config: configReducer,
});

const persistedReducer = persistReducer(persistConfig, reducers);

const middleware = [...getDefaultMiddleware(), thunk];

export const store = configureStore({
    reducer: persistedReducer,
    middleware,
});

export const persistor = persistStore(store);
