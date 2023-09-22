import { Route, Routes } from "react-router-dom";
import PublicRoutes from "./routes/PublicRoutes";
import UserRoutes from "./routes/UserRoutes";
import { useState, useEffect } from "react";

import { auth } from "./firebase";

// eslint-disable-next-line
import i18next from "./services/i18next";

const App = () => {
  const [user, setUser] = useState(null);

  // Use useEffect to set up an authentication listener
  useEffect(() => {
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
    <Routes>
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
    </Routes>
  );
};

export default App;
