import i18next from "i18next";
import { initReactI18next } from "react-i18next";

import en from "../locales/en.json";
import ar from "../locales/ar.json";

export const resources = {
    en: {
        translation: en,
    },
    ar: {
        translation: ar,
    },
};

i18next.use(initReactI18next).init({
    resources,
    lng: "en",
    fallbackLng: "ar",
    compatibilityJSON: "v3",
    interpolation: {
        escapeValue: false,
    },
});

export default i18next;