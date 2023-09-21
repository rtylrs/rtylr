import { createSlice } from '@reduxjs/toolkit';

export const alertSlice = createSlice({
    name: 'alert',
    initialState: {
        message: "",
        type: "success",
        show: false,
    },
    reducers: {
        setAlert: (state, actions) => {
            state.message = actions.payload.message;
            state.type = actions.payload.type;
            state.show = true;
        },

        closeAlert: (state) => {
            state.show = false;
        },
    },
});

export const {
    setAlert,
    closeAlert,
} = alertSlice.actions;

export default alertSlice.reducer;