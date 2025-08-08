import React from 'react';
import { Route, Routes } from 'react-router-dom';

import AuthRoutes from './routes/AuthRoutes';

function App() {
  return (
    <React.Fragment>
      <Routes>
        {AuthRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
      </Routes>
    </React.Fragment>
  );
}

export default App;
