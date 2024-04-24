//TODO: fix this

import { useEffect, useState } from "react";
import { Outlet, Navigate } from "react-router-dom";
import Loader from "../layouts/loader/Loader";

function BusinessRoute() {
  const [loading, setLoading] = useState(true);
  const [isBusiness, SetIsBusiness] = useState(true);

  useEffect(() => {
    SetIsBusiness(true);
    setLoading(false);
  }, []);

  if (!loading) {
    return isBusiness ? <Outlet /> : <Navigate to="/" />;
  } else {
    return <Loader />;
  }
}

export { BusinessRoute };
