import { useEffect } from 'react';
import { Route, Routes } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';

//Routes
import AuthRoutes from './routes/AuthRoutes';

// Redux imports
import { getUserInfo, selectIsAuthenticated } from './redux/reducers/userSlice';
import { RootState, AppDispatch } from './redux/store';

//Components
import HelmetContainer from './components/HelmetContainer';

//Hooks

function App() {
  const dispatch = useDispatch<AppDispatch>();

  const isAuthenticated = useSelector((state: RootState) =>
    selectIsAuthenticated(state)
  );

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(getUserInfo());
    }
  }, [dispatch, isAuthenticated]);

  return (
    <>
      <HelmetContainer />

      <Routes>
        {AuthRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
      </Routes>
    </>
  );
}

export default App;
