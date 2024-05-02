// Toastify.js

import { toast } from "react-toastify";

import "./Toastify.css"; // Import your custom CSS file

const success = (message) => {
    toast.success(message, {
        // position: toast.POSITION.BOTTOM_RIGHT,
        // autoClose: true, // No timeout
    });
};

const error = (message) => {
    toast.error(message, {
        // position: toast.POSITION.BOTTOM_RIGHT,
        // autoClose: true, // No timeout
    });
};

const notify = {
    success,
    error,
};

export default notify;