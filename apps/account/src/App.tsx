import React from 'react';
import { Route, Routes } from 'react-router-dom';

import AccountRoutes from './routes/AccountRoutes';

function App() {
  return (
    <React.Fragment>
      <Routes>
        {AccountRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
      </Routes>
    </React.Fragment>
  );
}

export default App;
