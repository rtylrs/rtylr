import { Route, Routes } from "react-router-dom";
import PublicRoutes from "./routes/PublicRoutes";

const App = () => {
  return (
    <Routes>
      {PublicRoutes.map((route, index) => {
        return (
          <Route
            key={index}
            path={route.path}
            exact={route.exact}
            element={route.element}
          />
        );
      })}
    </Routes>
  );
};

export default App;
