import api from '../../../lib/api';
import { API_ROUTES } from './routes';

const getMenu = async ({ id }) => {
  const res = await api.get(API_ROUTES.GET.GET_MENU_WITH_ID(id), {
    params: {
      detailed: true,
    },
  });
  return res.data;
};

const getMenuWithMenuId = async ({ id }) => {
  const res = await api.get(API_ROUTES.GET.GET_MENU_WITH_MENU_ID(id), {
    params: {
      detailed: true,
    },
  });
  return res.data;
};

const menuApi = {
  getMenu,
  getMenuWithMenuId,
};

export default menuApi;
