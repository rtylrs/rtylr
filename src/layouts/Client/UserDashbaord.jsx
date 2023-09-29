import React from "react";
import Navbar from "./UserNavbar";
import SideBar from "./UserSidebar";
import { auth } from "../../firebase";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";

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

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    // <AnimatedRoutes>
    <div className="Admincontainer">
      <Navbar sidebarOpen={sidebarOpen} openSidebar={openSidebar} />
      <div className="InnerContainerUser">{props.children}</div>
      <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
    </div>
    // </AnimatedRoutes>
  );
}

UserDashboard.propTypes = {
  children: PropTypes.node.isRequired,
};

export default UserDashboard;
