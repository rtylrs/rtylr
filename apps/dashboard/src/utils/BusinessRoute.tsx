import { Outlet, Navigate } from 'react-router-dom';
import { selectRole } from '../redux/reducers/userSlice';
import { useSelector } from 'react-redux';

function BusinessRoute() {
  const role = useSelector(selectRole);

  // return role === "business" || role === "admin" ? (
  //   <Outlet />
  // ) : (
  //   <Navigate to="/" />
  // );

  return <Outlet />;
}

export { BusinessRoute };
