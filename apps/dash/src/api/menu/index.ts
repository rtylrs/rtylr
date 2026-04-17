import api from '../../shared/lib/api';
import { API_ROUTES } from './routes';

const getMenu = async ({ id }: { id: string | number }) => {
  const res = await api.get(API_ROUTES.GET.GET_MENU_WITH_ID(id), {
    params: {
      detailed: true,
    },
  });
  return res.data;
};

const getMenuWithMenuId = async ({ id }: { id: string | number }) => {
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
