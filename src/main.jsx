import React, { Suspense } from "react";
import ReactDOM from "react-dom/client";
import App from "./App.jsx";
import "./styles/index.scss";
import { BrowserRouter } from "react-router-dom";
import Loader from "./layouts/loader/Loader.jsx";
import store from "./redux/store.js";
import { Provider } from "react-redux";
import { I18nextProvider } from "react-i18next";
import i18next from "./services/i18next";
import ScrollToTop from "./components/ScrollToTop";

import "react-lazy-load-image-component/src/effects/blur.css";

import { createTheme } from "@mui/material/styles";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";

const theme = createTheme({
  palette: {
    primary: {
      main: "#846552", // Your desired primary color
    },
  },
});

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <Suspense fallback={<Loader />}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <I18nextProvider i18n={i18next}>
          <Provider store={store}>
            <BrowserRouter>
              <ScrollToTop />
              <App />
            </BrowserRouter>
          </Provider>
        </I18nextProvider>
      </ThemeProvider>
    </Suspense>
  </React.StrictMode>
);
