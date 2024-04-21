import AsyncStorage from "@react-native-async-storage/async-storage";
import axios from "axios";
import { ASYNC_STORAGE_KEYS } from "../constants/async_storage_keys";
import i18next from "i18next";

const api = axios.create({
  // baseURL: "http://192.168.1.9:5000/api/v1/app",
});

api.interceptors.request.use(
  async (config) => {
    try {
      const user = await AsyncStorage.getItem(ASYNC_STORAGE_KEYS.AUTH_TOKEN);
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
