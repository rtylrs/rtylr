import { combineReducers, configureStore } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import userReducer from './reducers/userSlice';
import productsReducer from './reducers/productsSlice';
import cartReducer from './reducers/cartSlice';
import configReducer from './reducers/configSlice';

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

export const store = configureStore({
    reducer: persistedReducer,
});

export const persistor = persistStore(store);
