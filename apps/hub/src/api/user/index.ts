import api from '../../shared/lib/api';
import { API_ROUTES } from './routes';

const editAddons = async ({ id, addons }) => {
  try {
    const res = await api.patch(API_ROUTES.PATCH.EDIT_ADDON(id), {
      addons,
    });
    return res.data;
  } catch (error) {
    console.error('Error editing addons:', error);
    throw error;
  }
};

const userApi = {
  editAddons,
};

export default userApi;
