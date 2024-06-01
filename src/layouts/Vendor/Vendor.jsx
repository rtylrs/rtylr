import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { useSelector } from "react-redux";
import HelmetContainer from "../../components/HelmetContainer";
import { selectRole } from "../../redux/reducers/userSlice";
import FreeAccount from "../../components/FreeAccount";

function Vendor(props) {
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = React.useState(false);
  const user = useSelector((state) => state.user);
  const role = useSelector(selectRole);

  const openSidebar = () => {
    setSidebarOpen(true);
  };

  const closeSidebar = () => {
    setSidebarOpen(false);
  };

  React.useEffect(() => {
    const token = localStorage.getItem("token");

    if (!token) {
      navigate("/auth/login");
    }

    const url = new URL(window.location.href);
    if (!url.searchParams.has("userId")) {
      url.searchParams.append("userId", user?.userId);
      if (role !== "admin")
        url.searchParams.append("plan", user?.user?.planType);
      else url.searchParams.append("plan", "admin");
      window.history.replaceState({}, "", url.toString());
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <HelmetContainer title={`${user?.user?.name}'s Dashboard`} />
      {user?.user?.planType === "free" && <FreeAccount />}
      <div className="VendorContainer">
        <Navbar sidebarOpen={sidebarOpen} openSidebar={openSidebar} />
        <div className="InnerContainer">{props.children}</div>
        <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
      </div>
    </>
  );
}

Vendor.propTypes = {
  children: PropTypes.node.isRequired,
};

export default Vendor;
