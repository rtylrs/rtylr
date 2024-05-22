import api from "../../lib/api";
import { API_ROUTES } from "./routes";

const getAll = async () => {
  const res = await api.get(API_ROUTES.GET.GET_ALL);
  return res.data;
};

const UserCount = async () => {
  const res = await api.get(API_ROUTES.GET.GET_USER_COUNT);
  return res.data;
};

const BussinessCount = async () => {
  const res = await api.get(API_ROUTES.GET.GET_BUSSINESS_COUNT);
  return res.data;
};

const GetNews = async () => {
  const res = await api.get(API_ROUTES.GET.GET_NEWS);
  return res.data;
};


const adminApi = {
  getAll,
  UserCount,
  BussinessCount,
  GetNews
};

export default adminApi;
