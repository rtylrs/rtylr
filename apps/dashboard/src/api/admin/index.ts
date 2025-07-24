import http from "../../lib/adminApi";
import { API_ROUTES } from "./routes";

const getAllUsers = async () => {
  const res = await http.get(API_ROUTES.GET.GET_ALL_USERS);
  return res.data;
};

const getUserCount = async () => {
  const res = await http.get(API_ROUTES.GET.GET_USER_COUNT);
  return res.data;
};

const getBusinessCount = async () => {
  const res = await http.get(API_ROUTES.GET.GET_BUSINESS_COUNT);
  return res.data;
};

const addFaq = async (faqData) => {
  const res = await http.post(API_ROUTES.POST.ADD_FAQ, faqData);
  return res.data;
};

const adminApi = {
  getAllUsers,
  getUserCount,
  getBusinessCount,
  addFaq,
};

export default adminApi;
