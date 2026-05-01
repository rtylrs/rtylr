import api from '../../../lib/api';
import { API_ROUTES } from './routes';

const getAll = async () => {
  const res = await api.get(API_ROUTES.GET.GET_ALL);
  return res.data;
};

const businessApi = {
  getAll,
};

export default businessApi;
