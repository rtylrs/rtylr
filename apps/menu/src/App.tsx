import React, { Suspense, useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import PublicRoutes from './routes/PublicRoutes';
import { useConfigStore } from './stores/configStore';
import '@voxire/menu/styles';

const App = (): JSX.Element => {
  const location = useLocation();
  const fetchConfigData = useConfigStore((s) => s.fetchConfigData);

  useEffect(() => {
    fetchConfigData();
  }, []);

  return (
    <Suspense fallback={<FullScreenLoader />}>
      <Routes key={location.pathname} location={location}>
        {PublicRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
      </Routes>
    </Suspense>
  );
};

export default App;
