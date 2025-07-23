import i18next from "i18next";
import { initReactI18next } from "react-i18next";

import en from "../locales/en.json";
import fr from "../locales/fr.json";
import ar from "../locales/ar.json";

export const resources = {
  en: {
    translation: en,
  },
  fr: {
    translation: fr,
  },
  ar: {
    translation: ar,
  },
};

i18next.use(initReactI18next).init({
  resources,
  lng: "en",
  fallbackLng: "fr",
  compatibilityJSON: "v3",
  interpolation: {
    escapeValue: false,
  },
});

export const selectLanguage = (language) => {
  if (language) {
    i18next.changeLanguage(language);
    localStorage.setItem("language", language);
  } else {
    language = "en";
    i18next.changeLanguage(language);
    localStorage.setItem("language", language);
  }
};

export default i18next;
