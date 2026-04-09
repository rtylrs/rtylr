import React, { useEffect, Suspense } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './screens/Extras/NoMatch';
import { useUserStore } from './stores/userStore';
import Login from './screens/Login';

const App = (): JSX.Element => {
  const location = useLocation();
  const { isAuthenticated, getUserInfo, token } = useUserStore();

  useEffect(() => {
    if (token) {
      getUserInfo().catch(() => {});
    }
  }, []);

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <Suspense fallback={null}>
      <Routes key={location.pathname} location={location}>
        {AppRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
        <Route path="*" element={<NoMatch />} />
      </Routes>
    </Suspense>
  );
};

export default App;
