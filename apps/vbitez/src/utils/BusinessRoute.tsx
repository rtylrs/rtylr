import { Outlet, Navigate } from 'react-router-dom';
import { useUserStore } from '../stores/userStore';

function BusinessRoute() {
  const role = useUserStore((state) => state.role);

  // return role === "business" || role === "admin" ? (
  //   <Outlet />
  // ) : (
  //   <Navigate to="/" />
  // );

  return <Outlet />;
}

export { BusinessRoute };
