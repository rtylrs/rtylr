import store from "../redux/store";
import { closeAlert, setAlert } from "../redux/reducers/alertSlice";


const ShowAlert = (message, type) => {
    store.dispatch(setAlert({ message, type }));
    setTimeout(() => {
        store.dispatch(closeAlert());
    }, 5500);
};

export default ShowAlert;