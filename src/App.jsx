import { Route, Routes, useLocation } from "react-router-dom";

import PublicRoutes from "./routes/PublicRoutes";
import UserRoutes from "./routes/UserRoutes";
import AdminRoutes from "./routes/AdminRoutes";
import AuthRoutes from "./routes/AuthRoutes";
import VendorRoutes from "./routes/VendorRoutes";

import { useState, useEffect } from "react";

import { auth } from "./firebase";

// eslint-disable-next-line
import i18next from "./services/i18next";

import AlertModal from "./components/Alert/AlertModal";
import { AdminRoute } from "./utils/AdminRoute";
import { BusinessRoute } from "./utils/BusinessRoute";

const App = () => {
  const [user, setUser] = useState(null);
  const location = useLocation();

  // Use useEffect to set up an authentication listener
  useEffect(() => {
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
  }, []);

  return (
    <>
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
      </Routes>
    </>
  );
};

export default App;
