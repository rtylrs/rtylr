import { createSlice } from '@reduxjs/toolkit';
import { auth } from '../../firebase';

const initialState = {
    user: null,
    isAuthenticated: false,
};

const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {
        login: (state, action) => {
            state.user = action.payload;
            state.isAuthenticated = true;
        },
        logout: (state) => {
            auth
                .signOut()
                .then(() => {
                    state.user = null;
                    state.isAuthenticated = false;
                })
                .catch((error) => {
                    console.error('Error signing out:', error);
                });
        },
    },
});

export const { login, logout } = userSlice.actions;

export const selectUser = (state) => state.user.user;
export const selectIsAuthenticated = (state) => state.user.isAuthenticated;

export default userSlice.reducer;
