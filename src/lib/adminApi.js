import axios from "axios";
import i18next from "i18next";

const getBaseUrl = () => {
    const mode = localStorage.getItem("mode") || "prod";
    return mode === "dev"
        ? "http://localhost:1024/api/v1/admin"
        : "https://onion-e6t0.onrender.com/api/v1/admin";
};

const api = axios.create({
    baseURL: getBaseUrl(),
});

api.interceptors.request.use(
    async (config) => {
        try {
            const user =
                localStorage.getItem("userToken") ||
                sessionStorage.getItem("userToken");
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
