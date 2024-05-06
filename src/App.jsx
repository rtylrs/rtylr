import { Route, Routes, useLocation } from "react-router-dom";

import PublicRoutes from "./routes/PublicRoutes";
import UserRoutes from "./routes/UserRoutes";
import AdminRoutes from "./routes/AdminRoutes";
import AuthRoutes from "./routes/AuthRoutes";
import VendorRoutes from "./routes/VendorRoutes";

import { useState, useEffect } from "react";

// eslint-disable-next-line
import i18next from "./services/i18next";

import AlertModal from "./components/Alert/AlertModal";
import { AdminRoute } from "./utils/AdminRoute";
import { BusinessRoute } from "./utils/BusinessRoute";
import NoMatch from "./screens/Public/NoMatch";
import { useDispatch, useSelector } from "react-redux";
import { getAllProducts } from "./redux/reducers/productsSlice";
import { Helmet } from "react-helmet";
import { fetchConfigData } from "./redux/reducers/configSlice";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faChevronUp } from "@fortawesome/free-solid-svg-icons";
import useColors from "./hooks/useColors";
import { motion } from "framer-motion";
import { faWhatsapp } from "@fortawesome/free-brands-svg-icons";
import Maintanance from "./screens/Public/Maintanance";

const App = () => {
  const [user, setUser] = useState(null);
  const location = useLocation();
  const { primary } = useColors();
  const dispatch = useDispatch();
  const config = useSelector((state) => state.config);
  const [showTopBtn, setShowTopBtn] = useState(false);

  useEffect(() => {
    window.addEventListener("scroll", () => {
      if (window.scrollY > 400) {
        setShowTopBtn(true);
      } else {
        setShowTopBtn(false);
      }
    });
  }, []);

  useEffect(() => {
    // get config data
    dispatch(fetchConfigData());
    // get all products
    dispatch(getAllProducts());

    // get language from local storage
    const language = localStorage.getItem("language");
    if (language) {
      i18next.changeLanguage(language);
    }

    setUser({});

    //eslint-disable-next-line
  }, []);

  return (
    <>
      {config && (
        <Helmet>
          <title>Vbitez - Contactless QR Menu</title>
        </Helmet>
      )}

      {config.maintanance && !location.pathname.includes("/manage/") && (
        <Maintanance />
      )}

      {!location.pathname.includes("/restaurant/") && (
        <motion.div
          className="whatsappBtn"
          onClick={() => {
            window.open(
              "https://api.whatsapp.com/send?phone=96171642565&text=Hello%20Voxire",
              "_blank"
            );
          }}
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.2 }}
          whileTap={{ scale: 0.8 }}
          style={{
            right: showTopBtn ? "65px" : "20px",
          }}
        >
          <FontAwesomeIcon icon={faWhatsapp} />
        </motion.div>
      )}

      {showTopBtn && !location.pathname.includes("/restaurant/") && (
        <motion.div
          className="ScrollToTop"
          onClick={() => {
            setShowTopBtn(false);
            window.scrollTo({
              top: 0,
              behavior: "smooth",
            });
          }}
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.2 }}
          whileTap={{ scale: 0.8 }}
          style={{
            backgroundColor: primary,
          }}
        >
          <FontAwesomeIcon icon={faChevronUp} />
        </motion.div>
      )}

      <AlertModal />

      <Routes key={location.pathname} location={location}>
        {/* Auth Routes */}
        {AuthRoutes.map((route, index) => (
          <Route
            key={index}
            path={route.path}
            exact={route.exact}
            element={route.element}
          />
        ))}

        {/* Public Routes */}
        {PublicRoutes.map((route, index) => (
          <Route
            key={index}
            path={route.path}
            exact={route.exact}
            element={route.element}
          />
        ))}

        {/* User routes */}
        {UserRoutes.map((route, index) => (
          <Route
            key={index}
            path={route.path}
            exact={route.exact}
            element={route.element}
          />
        ))}

        <Route exact path="/manage" element={<AdminRoute />}>
          {AdminRoutes.map((route, index) => (
            <Route
              key={index}
              path={route.path}
              exact={route.exact}
              element={route.element}
            />
          ))}
        </Route>

        <Route exact path="/restaurant" element={<BusinessRoute />}>
          {VendorRoutes.map((route, index) => (
            <Route
              key={index}
              path={route.path}
              exact={route.exact}
              element={route.element}
            />
          ))}
        </Route>

        <Route path="*" element={<NoMatch />} />
      </Routes>
    </>
  );
};

export default App;
