import { configureStore } from '@reduxjs/toolkit';
import alertReducer from './reducers/alertSlice';
import userReducer from './reducers/userSlice';
import productsReducer from './reducers/productsSlice';

// export default store from 'store.js';
export default configureStore({
    reducer: {
        customAlert: alertReducer,
        user: userReducer,
        products: productsReducer,
    },
});