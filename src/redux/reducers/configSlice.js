import { createSlice } from '@reduxjs/toolkit';
import api from '../../lib/api';

const initialState = {
    websiteData: null,
    aboutContent: null,
    maintanance: false,
    bussinessCount: 0,
    contactInfo: {
        contactEmail: null,
        contactNumber: null,
    },
    status: 'idle',
    error: null,
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
        },
        setBussinessCount: (state, action) => {
            state.bussinessCount = action.payload;
        },
        setContactInfo: (state, action) => {
            state.contactInfo = action.payload;
        }
    },
});

export const {
    setWebsiteData,
    setAboutContent,
    setMaintanance,
    setBussinessCount,
    setContactInfo
} = configSlice.actions;

export default configSlice.reducer;

export const fetchConfigData = () => async (dispatch) => {
    try {
        const response = await api.get('/config');
        const { maintanance, aboutContent, count, contactEmail, contactNumber } = response.data;
        dispatch(setAboutContent(aboutContent));
        dispatch(setMaintanance(maintanance));
        dispatch(setBussinessCount(count));
        dispatch(setContactInfo({
            contactEmail,
            contactNumber
        }));
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
