import React, { Suspense, useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { FullScreenLoader } from '@voxire/components';
import PublicRoutes from './routes/PublicRoutes';
import { fetchConfigData } from './redux/reducers/configSlice';
import { AppDispatch } from './redux/store';
import '@voxire/menu/styles';

const App = (): JSX.Element => {
  const location = useLocation();
  const dispatch = useDispatch<AppDispatch>();

  useEffect(() => {
    dispatch(fetchConfigData());
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
