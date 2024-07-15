import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { useSelector } from "react-redux";
import HelmetContainer from "../../components/HelmetContainer";
import { selectRole } from "../../redux/reducers/userSlice";
import FreeAccount from "../../components/FreeAccount";
import { RootState } from "../../redux/store";

const Vendor: React.FC<{
  children?: React.ReactNode;
}> = (props) => {
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = React.useState<boolean>(false);
  const user = useSelector((state: RootState) => state.user);
  const role = useSelector(selectRole);

  const openSidebar = (open: boolean) => {
    setSidebarOpen(open);
  };

  const closeSidebar = () => {
    setSidebarOpen(false);
  };

  React.useEffect(() => {
    const token = localStorage.getItem("token");
    const mode = localStorage.getItem("mode");

    if (!token) {
      navigate("/auth/login");
    }

    const url = new URL(window.location.href);
    if (!url.searchParams.has("userId")) {
      url.searchParams.append("userId", (user?.userId || "").substring(0, 6));
      if (role !== "admin") {
        url.searchParams.append("plan", user?.user?.planType || "");
      } else {
        url.searchParams.append("plan", "admin");
      }

      if (mode === "dev") {
        url.searchParams.append("mode", "dev");
      }

      window.history.replaceState({}, "", url.toString());
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigate, user?.userId, user?.user?.planType, role]);

  return (
    <>
      <HelmetContainer title={`${user?.user?.name || "User"}'s Dashboard`} />
      {user?.user?.planType === "free" && <FreeAccount />}
      <div className="VendorContainer">
        <Navbar sidebarOpen={sidebarOpen} openSidebar={openSidebar} />
        <div className="InnerContainer">{props.children}</div>
        <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
      </div>
    </>
  );
};

Vendor.propTypes = {
  children: PropTypes.node.isRequired,
};

export default Vendor;
