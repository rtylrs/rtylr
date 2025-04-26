import React from "react";
import Navbar from "./UserNavbar";
import SideBar from "./UserSidebar";
import { auth } from "../../firebase";
import { useNavigate } from "react-router-dom";

function UserDashboard(props) {
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = React.useState(false);

  const openSidebar = () => {
    setSidebarOpen(true);
  };

  const closeSidebar = () => {
    setSidebarOpen(false);
  };

  React.useEffect(() => {
    // if not user redirect to login
    auth.onAuthStateChanged((user) => {
      if (!user) {
        navigate("/");
      }
    });
  }, []);

  return (
    // <AnimatedRoutes>
    <div className="Admincontainer">
      <Navbar sidebarOpen={sidebarOpen} openSidebar={openSidebar} />
      <div className="InnerContainer">{props.children}</div>
      <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
    </div>
    // </AnimatedRoutes>
  );
}

export default UserDashboard;
