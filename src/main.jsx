/*
  File Created By: Vite js
  Last Edited: Sunday 12 May 2024
  Last Edited by: Mohamad Homsi
*/

import React, { Suspense } from "react";
import ReactDOM from "react-dom/client";
import App from "./App.jsx";
import "./styles/index.scss";
import { BrowserRouter } from "react-router-dom";
import Loader from "./layouts/loader/Loader.jsx";
import { store, persistor } from "./redux/store.js";
import { Provider } from "react-redux";
import { I18nextProvider } from "react-i18next";
import i18next from "./services/i18next";
import ScrollToTop from "./components/ScrollToTop";

import "react-lazy-load-image-component/src/effects/blur.css";

import { createTheme } from "@mui/material/styles";
import { ThemeProvider } from "@mui/material/styles";
import { SidebarProvider } from "./HOCs/SidebarProvider.jsx";
import { ToastContainer } from "react-toastify";
import { CookiesProvider } from "react-cookie";
import { PersistGate } from "redux-persist/integration/react";

const theme = createTheme({
  palette: {
    primary: {
      main: "#0a102f",
    },
  },
});

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <CookiesProvider>
      <ThemeProvider theme={theme}>
        <Suspense fallback={<Loader />}>
          <I18nextProvider i18n={i18next}>
            <Provider store={store}>
              <PersistGate loading={null} persistor={persistor}>
                <BrowserRouter>
                  <SidebarProvider>
                    <ScrollToTop />
                    <App />
                    <ToastContainer />
                  </SidebarProvider>
                </BrowserRouter>
              </PersistGate>
            </Provider>
          </I18nextProvider>
        </Suspense>
      </ThemeProvider>
    </CookiesProvider>
  </React.StrictMode>
);
