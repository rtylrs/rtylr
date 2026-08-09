import api from '../../../lib/api';
import { API_ROUTES } from './routes';

const getNews = async () => {
  const res = await api.get(API_ROUTES.GET.GET_NEWS);
  return res.data;
};

const getSpecificNews = async ({ id }) => {
  const res = await api.get(API_ROUTES.GET.GET_NEWS_WITH_ID(id));
  return res.data;
};

const configApi = {
  getNews,
  getSpecificNews,
};

export default configApi;
