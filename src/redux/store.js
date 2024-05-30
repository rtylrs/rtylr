import { combineReducers, configureStore } from "@reduxjs/toolkit";
import { persistReducer, persistStore } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import userReducer from './reducers/userSlice';
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
