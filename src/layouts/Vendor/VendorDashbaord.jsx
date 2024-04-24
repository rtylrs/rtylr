import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import PropTypes from "prop-types";

function VendorDashboard(props) {
  const [sidebarOpen, setSidebarOpen] = React.useState(false);

  const openSidebar = () => {
    setSidebarOpen(true);
  };

  const closeSidebar = () => {
    setSidebarOpen(false);
  };

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
