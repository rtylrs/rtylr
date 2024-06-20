/*
  File Created By: Mohamad Homsi
  Last Edited: Sunday 12 May 2024
  Last Edited by: Mohamad Homsi
*/

import { Outlet, Navigate } from "react-router-dom";
import { selectRole } from "../redux/reducers/userSlice";
import { useSelector } from "react-redux";

function AdminRoute() {
  const role = useSelector(selectRole);

  return role === "admin" ? <Outlet /> : <Navigate to="/" />;
}

export { AdminRoute };
