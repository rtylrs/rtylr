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

import i18next from "./shared/services/i18next";

//Components
import Cookies from "./shared/components/Cookies/Cookies";
import Maintanance from "./features/extras/Maintanance";
import NoMatch from "./features/extras/NoMatch";

//Utilities
import { getCookie } from "./shared/lib/cookies";
import HelmetContainer from "./shared/components/HelmetContainer";

//Hooks
import useColors from "./shared/hooks/useColors";
import { useFullscreen } from "./shared/hooks/useFullscreen";

// Route Checks
import { BusinessRoute } from "./shared/utils/BusinessRoute";

// Store imports
import { useConfigStore } from "./shared/stores/configStore";
import { useUserStore } from "./shared/stores/userStore";
import { useGetConfig } from "./shared/hooks/queries/useConfig";

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

  const role = useUserStore((state) => state.role);

  useEffect(() => {
    const language = getCookie("language");
    if (language) {
      i18next.changeLanguage(language);
    }
  }, []);

  return (
    <React.Fragment>
      <HelmetContainer title={"Rtylr"} favicon={"/favicon.ico"} />

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
