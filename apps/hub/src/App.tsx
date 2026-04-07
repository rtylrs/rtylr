import React from "react";
//Package imports
import { useState, useEffect } from "react";
import { Route, Routes, useLocation } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Chart as ChartJS,
  LineElement,
  PointElement,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  CategoryScale,
} from "chart.js";

//Routes
import AppRoutes from "./routes/AppRoutes";

import i18next from "./services/i18next";

//Components
import Cookies from "./components/Cookies/Cookies";
import Maintanance from "./screens/Maintanance";
import NoMatch from "./screens/Extras/NoMatch";

//Utilities
import { getCookie } from "./lib/cookies";
import HelmetContainer from "./components/HelmetContainer";

//Hooks
import useColors from "./hooks/useColors";
import { useFullscreen } from "./hooks/useFullscreen";

// Route Checks
import { BusinessRoute } from "./utils/BusinessRoute";

// Store imports
import { useConfigStore } from "./stores/configStore";
import { useUserStore } from "./stores/userStore";
import { useGetConfig } from "./hooks/queries/useConfig";

ChartJS.register(
  LineElement,
  PointElement,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  CategoryScale
);

const App = (): JSX.Element => {
  //Hooks
  const location = useLocation();
  const { primary } = useColors();

  // Initialize fullscreen (will default to fullscreen on first run)
  useFullscreen();

  // Zustand stores
  const config = useConfigStore();
  const role = useUserStore((state) => state.role);

  // Fetch config data
  useGetConfig();

  useEffect(() => {
    const language = getCookie("language");
    if (language) {
      i18next.changeLanguage(language);
    }
  }, []);

  return (
    <React.Fragment>
      {config && <HelmetContainer title={"Rtylr"} favicon={"/favicon.ico"} />}

      {config.maintanance && <Maintanance />}

      <Cookies />

      <Routes key={location.pathname} location={location}>
        <Route path="/" element={<BusinessRoute />}>
          {AppRoutes.map((route, index) => (
            <Route key={index} path={route.path} element={route.element}>
              {route.children?.map((child, childIndex) => (
                <Route
                  key={childIndex}
                  index={child.index}
                  path={child.path}
                  element={child.element}
                />
              ))}
            </Route>
          ))}
        </Route>

        <Route path="*" element={<NoMatch />} />
      </Routes>
    </React.Fragment>
  );
};

export default App;
