import React from "react";
//Package imports
import { useState, useEffect } from "react";
import { Route, Routes, useLocation, Navigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { motion } from "framer-motion";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faChevronUp } from "@fortawesome/free-solid-svg-icons";
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
import AuthRoutes from "./routes/AuthRoutes";

import i18next from "./services/i18next";

//Screens
import NoMatch from "./screens/Public/Extras/NoMatch";
import Maintanance from "./screens/Public/Maintanance";

//Components
import Cookies from "./components/Cookies/Cookies";
import HelmetContainer from "./components/HelmetContainer";

//Hooks
import useColors from "./hooks/useColors";

// Other imports
import { fetchConfigData } from "./redux/reducers/configSlice";
import {
  getUserInfo,
  selectIsAuthenticated,
  selectRole,
} from "./redux/reducers/userSlice";

ChartJS.register(
  LineElement,
  PointElement,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  CategoryScale
);

function App() {
  const location = useLocation();
  const dispatch = useDispatch();
  const { primary } = useColors();
  
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const role = useSelector(selectRole);

  const [showScrollToTop, setShowScrollToTop] = useState(false);
  const [isMaintenanceMode, setIsMaintenanceMode] = useState(false);

  // Scroll to top functionality
  useEffect(() => {
    const handleScroll = () => {
      setShowScrollToTop(window.scrollY > 300);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  // Load configuration data
  useEffect(() => {
    dispatch(fetchConfigData());
  }, [dispatch]);

  // Load user info if authenticated
  useEffect(() => {
    if (isAuthenticated) {
      dispatch(getUserInfo());
    }
  }, [dispatch, isAuthenticated]);

  // Maintenance mode check
  useEffect(() => {
    const maintenanceMode = process.env.REACT_APP_MAINTENANCE_MODE === "true";
    setIsMaintenanceMode(maintenanceMode);
  }, []);

  if (isMaintenanceMode) {
    return <Maintanance />;
  }

  return (
    <>
      <HelmetContainer />
      
      <Routes>
        {/* Authentication Routes */}
        {AuthRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}

        {/* Default redirect to login */}
        <Route path="/" element={<Navigate to="/auth/login" replace />} />
        
        {/* 404 Not Found */}
        <Route path="*" element={<NoMatch />} />
      </Routes>

      {/* Scroll to Top Button */}
      {showScrollToTop && (
        <motion.button
          className="scroll-to-top"
          onClick={scrollToTop}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
          style={{
            position: "fixed",
            bottom: "20px",
            right: "20px",
            backgroundColor: primary,
            color: "white",
            border: "none",
            borderRadius: "50%",
            width: "50px",
            height: "50px",
            cursor: "pointer",
            zIndex: 1000,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <FontAwesomeIcon icon={faChevronUp} />
        </motion.button>
      )}

      <Cookies />
    </>
  );
}

export default App;
