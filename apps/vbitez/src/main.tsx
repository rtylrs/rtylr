import { Suspense } from "react";
import { injectTheme } from '@voxire/config';
import ReactDOM from "react-dom/client";
import App from "./App";
import { BrowserRouter } from "react-router-dom";
import { FullScreenLoader } from "./components/FullScreenLoader";
import { I18nextProvider } from "react-i18next";
import i18next from "./services/i18next";
import ScrollToTop from "./components/ScrollToTop";
import { Toaster } from "sonner";
import { CookiesProvider } from "react-cookie";
import Modal from "react-modal";

import "react-lazy-load-image-component/src/effects/blur.css";


Modal.setAppElement("#root");
injectTheme();

const rootElement = document.getElementById("root");

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <CookiesProvider>
        <Suspense fallback={<FullScreenLoader />}>
          <I18nextProvider i18n={i18next}>
            <BrowserRouter>
              <App />
              <Toaster expand={false} richColors />
              <ScrollToTop />
            </BrowserRouter>
          </I18nextProvider>
        </Suspense>
    </CookiesProvider>
  );
} else {
  <div>The App Failed To Load</div>;
}
