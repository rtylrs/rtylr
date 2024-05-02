import axios from "axios";
import i18next from "i18next";

const api = axios.create({
  baseURL: "http://localhost:1024/api/v1/app",
});

api.interceptors.request.use(
  async (config) => {
    try {
      const user = localStorage.getItem("userToken") || sessionStorage.getItem("userToken");
      if (user) {
        config.headers["authorization"] = `Bearer ${user}`;
      }
      config.headers["accept-language"] = i18next.language;
      return config;
    } catch (error) {
      return Promise.reject(error);
    }
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default api;
