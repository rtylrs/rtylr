import React from "react";
import Navbar from "./Navbar";
import SideBar from "./Sidebar";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { useSelector } from "react-redux";
import { selectRole } from "../../redux/reducers/userSlice";

function Admin(props) {
  const navigate = useNavigate();
  const role = useSelector(selectRole);
  const [sidebarOpen, setSidebarOpen] = React.useState(false);

  const openSidebar = () => {
    setSidebarOpen(true);
  };

  const closeSidebar = () => {
    setSidebarOpen(false);
  };

  React.useEffect(() => {
    // if not user redirect to login
    const token = localStorage.getItem("token");

    if (!token) {
      navigate("/auth/login");
    } else {
      if (role !== "admin") {
        navigate("/");
      }
    }
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

Admin.propTypes = {
  children: PropTypes.node.isRequired,
};

export default Admin;
