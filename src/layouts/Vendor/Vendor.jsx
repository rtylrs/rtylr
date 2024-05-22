import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { useSelector } from "react-redux";
import HelmetContainer from "../../components/HelmetContainer";
import { selectRole } from "../../redux/reducers/userSlice";

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
    } else {
      if (role !== "business") {
        navigate("/");
      }
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <HelmetContainer
        title={`${user?.user?.name}'s Dashboard`}
        favicon={user?.user?.favicon}
      />
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
