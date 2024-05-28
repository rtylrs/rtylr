import api from "../../lib/api";
import { API_ROUTES } from "./routes";

const GetNews = async () => {
    const res = await api.get(API_ROUTES.GET.GET_NEWS);
    return res.data;
};


const configApi = {
    GetNews
};

export default configApi;
