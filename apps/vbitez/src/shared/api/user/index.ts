import api from '../../lib/api';
import { API_ROUTES } from './routes';

const editAddons = async ({ id, addons }: { id: string | number; addons: unknown }) => {
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
