export const API_ROUTES = {
  GET: {},
  POST: {},
  PATCH: {
    EDIT_ADDON: (id: string | number) => `/business/${id}/enableAddons`,
  },
  DELETE: {},
};
