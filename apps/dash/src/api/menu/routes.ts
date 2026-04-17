export const API_ROUTES = {
  GET: {
    GET_MENU_WITH_ID: (id: string | number) => `/menu/${id}`,
    GET_MENU_WITH_MENU_ID: (id: string | number) => `/menu/menu/${id}`,
  },
  POST: {},
};
