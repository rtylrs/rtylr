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
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "./lib/react-query";

import './globals.css';
import "react-lazy-load-image-component/src/effects/blur.css";
import { initTheme } from './stores/themeStore';

// Apply persisted theme before first render to prevent flash
initTheme();

Modal.setAppElement("#root");
injectTheme();

const rootElement = document.getElementById("root");

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
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
    </QueryClientProvider>
  );
} else {
  <div>The App Failed To Load</div>;
}
