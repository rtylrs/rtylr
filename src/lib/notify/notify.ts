// Toastify.js

import { toast } from "react-toastify";

import "./Toastify.scss"; // Import your custom CSS file

const success = (message: string) => {
  toast.success(message, {
    // position: toast.POSITION.BOTTOM_RIGHT,
    // autoClose: true, // No timeout
  });
};

const error = (message: string) => {
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
