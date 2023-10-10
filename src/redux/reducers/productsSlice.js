import { createSlice } from '@reduxjs/toolkit';
import { db } from '../../firebase';

const initialState = {
    products: [],
    collections: [],
    selectedCollection: null,
};

export const productsSlice = createSlice({
    name: 'products',
    initialState,
    reducers: {
        setProducts: (state, action) => {
            state.products = action.payload;
        },
        setCollections: (state, action) => {
            state.collections = action.payload;
        },
        setSelectedCollection: (state, action) => {
            state.selectedCollection = action.payload;
        },
    },
});

export const { setProducts, setCollections, setSelectedCollection } = productsSlice.actions;

export const saveProduct = (product) => (dispatch) => {
    // Save product to database or API
    // ...
    // Dispatch setProducts action with updated products array
    // dispatch(setProducts(updatedProductsArray));
};

export const getProductsByCollection = (collectionId) => (dispatch) => {
    // Fetch products from database or API by collection ID
    // ...
    // Dispatch setProducts action with fetched products array
    // dispatch(setProducts(fetchedProductsArray));
};

export const getAllProducts = () => (dispatch) => {
    db.collection('Products').get().then((snapshot) => {
        const products = [];
        snapshot.forEach((doc) => {
            const product = Object.assign({ id: doc.id }, doc.data());
            products.push(product);
        });
        dispatch(setProducts(products));
    });
};

export default productsSlice.reducer;
