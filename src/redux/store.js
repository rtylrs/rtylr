import { configureStore } from '@reduxjs/toolkit';
import alertReducer from './reducers/alertSlice';

// export default store from 'store.js';
export default configureStore({
    reducer: {
        customAlert: alertReducer,
    },
});