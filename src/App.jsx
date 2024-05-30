/*
  File Created By: Mohamad Homsi
  Last Edited: Thursday 16 May 2024
  Last Edited by: Mohamad Homsi
*/

//Default imports
import { useState, useEffect } from "react";
import { Route, Routes, useLocation } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { motion } from "framer-motion";

//Routes
import PublicRoutes from "./routes/PublicRoutes";
import UserRoutes from "./routes/UserRoutes";
import AdminRoutes from "./routes/AdminRoutes";
import AuthRoutes from "./routes/AuthRoutes";
import VendorRoutes from "./routes/VendorRoutes";

// eslint-disable-next-line
import i18next from "./services/i18next";

//Screens
import NoMatch from "./screens/Public/Extras/NoMatch";
import Maintanance from "./screens/Public/Maintanance";

//Components
import Cookies from "./components/Cookies";
import HelmetContainer from "./components/HelmetContainer";

//Hooks
import useColors from "./hooks/useColors";

// Route Checks
import { AdminRoute } from "./utils/AdminRoute";
import { BusinessRoute } from "./utils/BusinessRoute";

// Other imports
import { getAllProducts } from "./redux/reducers/productsSlice";
import { fetchConfigData } from "./redux/reducers/configSlice";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faChevronUp } from "@fortawesome/free-solid-svg-icons";
import { faWhatsapp } from "@fortawesome/free-brands-svg-icons";

const App = () => {
  //Hooks
  const location = useLocation();
  const { primary } = useColors();
  const dispatch = useDispatch();

  // Redux
  const config = useSelector((state) => state.config);

  // States
  const [showTopBtn, setShowTopBtn] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 400) {
        setShowTopBtn(true);
      } else {
        setShowTopBtn(false);
      }
    };

    window.addEventListener("scroll", handleScroll);

    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);

  useEffect(() => {
    // get config data
    dispatch(fetchConfigData());

    // get all products
    dispatch(getAllProducts());

    const language = localStorage.getItem("language");
    if (language) {
      i18next.changeLanguage(language);
    }

    //eslint-disable-next-line
  }, []);

  return (
    <>
      {config && (
        <HelmetContainer
          title={"Vbitez - Contactless QR Menu"}
          favicon={"/favicon.ico"}
        />
      )}

      {config.maintanance && !location.pathname.includes("/manage/") && (
        <Maintanance />
      )}

      <Cookies />

      {!location.pathname.includes("/restaurant/") &&
        !location.pathname.includes("/app/") && (
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

        <Route exact path="/app/restaurant" element={<BusinessRoute />}>
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
