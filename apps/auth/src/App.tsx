import React from 'react';
import { Route, Routes } from 'react-router-dom';
import { ErrorBoundary } from '@voxire/ui';

import AuthRoutes from './routes/AuthRoutes';

function App() {
  return (
    <ErrorBoundary>
      <Routes>
        {AuthRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
      </Routes>
    </ErrorBoundary>
  );
}

export default App;
