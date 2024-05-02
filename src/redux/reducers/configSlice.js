import { createSlice } from '@reduxjs/toolkit';
import api from '../../lib/api';

const initialState = {
    websiteData: null,
    aboutContent: null,
    maintanance: false,
    status: 'idle',
    error: null
};

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
        setMaintanance: (state, action) => {
            state.maintanance = action.payload;
        }
    },
});

export const {
    setWebsiteData,
    setAboutContent,
    setMaintanance
} = configSlice.actions;

export default configSlice.reducer;

export const fetchConfigData = () => async (dispatch) => {
    console.log('Fetching config data...');
    try {
        const response = await api.get('/config');
        const { maintanance, aboutContent } = response.data;
        dispatch(setAboutContent(aboutContent));
        dispatch(setMaintanance(maintanance));
    } catch (error) {
        console.error('Error fetching config data:', error);
    }
};

export const updateAboutContent = (data) => async (dispatch) => {
    try {
        const response = await api.put('/config/update', data);
        dispatch(setAboutContent(response.data.aboutContent));
    } catch (error) {
        console.error('Error updating about content:', error);
    }
};
