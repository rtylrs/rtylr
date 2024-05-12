import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import { auth } from "../../firebase";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";

function VendorDashboard(props) {
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
    // auth.onAuthStateChanged((user) => {
    //   if (!user) {
    //     navigate("/");
    //   }
    // });
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

VendorDashboard.propTypes = {
  children: PropTypes.node.isRequired,
};

export default VendorDashboard;
