import { combineReducers, configureStore, getDefaultMiddleware } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import userReducer from './reducers/userSlice';
import productsReducer from './reducers/productsSlice';
import cartReducer from './reducers/cartSlice';
import configReducer from './reducers/configSlice';
import menuReducer from './reducers/menuSlice';
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
    menu: menuReducer
});

const persistedReducer = persistReducer(persistConfig, reducers);


export const store = configureStore({
    reducer: persistedReducer,
    middleware: (getDefaultMiddleware) => getDefaultMiddleware({
        serializableCheck: false
    }).concat(thunk)
});

export const persistor = persistStore(store);
