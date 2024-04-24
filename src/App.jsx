import { Link, Route, Routes, useLocation } from "react-router-dom";

import PublicRoutes from "./routes/PublicRoutes";
import UserRoutes from "./routes/UserRoutes";
import AdminRoutes from "./routes/AdminRoutes";
import AuthRoutes from "./routes/AuthRoutes";
import VendorRoutes from "./routes/VendorRoutes";

import { useState, useEffect } from "react";

import { auth, db } from "./firebase";

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

const App = () => {
  const [user, setUser] = useState(null);
  const location = useLocation();
  const { primary } = useColors();
  const dispatch = useDispatch();
  const config = useSelector((state) => state.config.websiteData);
  const [showTopBtn, setShowTopBtn] = useState(false);

  useEffect(() => {
    window.addEventListener("scroll", () => {
      if (window.scrollY > 400) {
        // 400 is the height of the button
        setShowTopBtn(true); // Show the button
      } else {
        setShowTopBtn(false); // Hide the button
      }
    });
  }, []);

  // Use useEffect to set up an authentication listener
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

    const unsubscribe = auth.onAuthStateChanged((authUser) => {
      if (authUser) {
        // User is logged in
        setUser(authUser);
      } else {
        // User is not logged in
        setUser(null);
      }
    });

    return () => {
      // Unsubscribe from the listener when the component unmounts
      unsubscribe();
    };

    //eslint-disable-next-line
  }, []);

  return (
    <>
      {config && (
        <Helmet>
          <title>Vbitez - Contactless QR Menu</title>
        </Helmet>
      )}

      {showTopBtn && !location.pathname.includes("/restaurant/") && (
        <div
          className="ScrollToTop"
          onClick={() => {
            window.scrollTo({
              top: 0,
              behavior: "smooth",
            });
          }}
          style={{
            backgroundColor: primary,
          }}
        >
          <FontAwesomeIcon icon={faChevronUp} />
        </div>
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

        {/* Conditional rendering for user routes */}
        {user && (
          <>
            {UserRoutes.map((route, index) => (
              <Route
                key={index}
                path={route.path}
                exact={route.exact}
                element={route.element}
              />
            ))}
          </>
        )}

        <Route exact path="/manage" element={<AdminRoute />}>
          {
            // Admin routes
            user && (
              <>
                {AdminRoutes.map((route, index) => (
                  <Route
                    key={index}
                    path={route.path}
                    exact={route.exact}
                    element={route.element}
                  />
                ))}
              </>
            )
          }
        </Route>

        <Route exact path="/business" element={<BusinessRoute />}>
          {
            // Vendor routes
            user && (
              <>
                {VendorRoutes.map((route, index) => (
                  <Route
                    key={index}
                    path={route.path}
                    exact={route.exact}
                    element={route.element}
                  />
                ))}
              </>
            )
          }
        </Route>

        <Route path="*" element={<NoMatch />} />
      </Routes>
    </>
  );
};

export default App;
