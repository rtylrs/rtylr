import { createSlice } from '@reduxjs/toolkit';
import { db } from '../../firebase';


// Define the initial state of the slice
const initialState = {
    websiteData: null,
    aboutContent: null,
    status: 'idle',
    error: null
};

// Create the slice
const configSlice = createSlice({
    name: 'config',
    initialState,
    reducers: {
        setWebsiteData: (state, action) => {
            state.websiteData = action.payload;
        },
        setAboutContent: (state, action) => {
            state.aboutContent = action.payload;
        },
    },
});

// Export actions
export const {
    setWebsiteData
    , setAboutContent
} = configSlice.actions;

// Export the reducer
export default configSlice.reducer;

export const fetchConfigData = () => (dispatch) => {
    db.collection('config').get().then((snapshot) => {
        snapshot.forEach((doc) => {
            const config = Object.assign({ id: doc.id }, doc.data());

            if (config.id === 'websiteData') {
                dispatch(setWebsiteData(config));
            }

            if (config.id === 'about') {
                dispatch(setAboutContent(config));
            }
        });
    });
};


export const UpdateAboutContent = (data) => (dispatch) => {
    db.collection('config').doc('about').update(data).then(() => {
        dispatch(setAboutContent(data));
    });
};