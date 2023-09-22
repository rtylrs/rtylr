import { configureStore } from '@reduxjs/toolkit';
import alertReducer from './reducers/alertSlice';
import userReducer from './reducers/userSlice';

// export default store from 'store.js';
export default configureStore({
    reducer: {
        customAlert: alertReducer,
        user: userReducer,
    },
});