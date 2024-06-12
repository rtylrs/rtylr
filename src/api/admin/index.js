import adminApi from "../../lib/adminApi";
import { API_ROUTES } from "./routes";

const getAll = async () => {
  const res = await adminApi.get(API_ROUTES.GET.GET_ALL);
  return res.data;
};

const UserCount = async () => {
  const res = await adminApi.get(API_ROUTES.GET.GET_USER_COUNT);
  return res.data;
};

const BussinessCount = async () => {
  const res = await adminApi.get(API_ROUTES.GET.GET_BUSSINESS_COUNT);
  return res.data;
};

const addFaq = async (faqData) => {
  const res = await adminApi.post(API_ROUTES.POST.ADD_FAQ, faqData);
  return res.data;
};

const adminApiExport = {
  getAll,
  UserCount,
  BussinessCount,
  addFaq,
};

export default adminApiExport;
