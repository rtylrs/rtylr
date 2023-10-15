import { createSlice } from '@reduxjs/toolkit';
import { db } from '../../firebase';


// Define the initial state of the slice
const initialState = {
    configData: null,
    status: 'idle',
    error: null
};

// Create the slice
const configSlice = createSlice({
    name: 'config',
    initialState,
    reducers: {
        setConfig: (state, action) => {
            state.configData = action.payload;
        },
    },
});

// Export actions
export const { setConfig } = configSlice.actions;

// Export the reducer
export default configSlice.reducer;

export const fetchConfigData = () => (dispatch) => {
    db.collection('config').get().then((snapshot) => {
        const configData = [];
        snapshot.forEach((doc) => {
            const config = Object.assign({ id: doc.id }, doc.data());
            configData.push(config);
        });
        dispatch(setConfig(configData));
    });
};
