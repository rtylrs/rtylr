import { auth, db } from "../firebase";
import { useEffect, useState } from "react";
import { Outlet, Navigate } from "react-router-dom";
import Loader from "../layouts/loader/Loader";

function AdminRoute() {
  const [loading, setLoading] = useState(true);
  const [admin, setIsAdmin] = useState(true);

  useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged((user) => {
      if (user) {
        db.collection("Users")
          .doc(auth.currentUser?.uid)
          .get()
          .then((doc) => {
            if (!doc.data().isAdmin) {
              setIsAdmin(false);
            }
            setLoading(false);
          });
      } else {
        // Handle the case when the user is not authenticated
        setLoading(false);
      }
    });

    return () => {
      // Unsubscribe from the auth state listener when the component unmounts
      unsubscribe();
    };
  }, []);

  if (!loading) {
    return admin ? <Outlet /> : <Navigate to="/" />;
  } else {
    // Optionally, you can render a loading indicator here while checking the admin status
    return <Loader />;
  }
}

export { AdminRoute };
