import { Outlet, Navigate } from "react-router-dom";
import { selectRole } from "../redux/reducers/userSlice";
import { useSelector } from "react-redux";

function AdminRoute() {
  const role = useSelector(selectRole);

  return role === "admin" ? <Outlet /> : <Navigate to="/" />;
}

export { AdminRoute };
