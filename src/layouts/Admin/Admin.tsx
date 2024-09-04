import React from "react";
import Navbar from "./Navbar";
import SideBar from "./Sidebar";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { useSelector } from "react-redux";
import { selectRole } from "../../redux/reducers/userSlice";
import HelmetContainer from "../../components/HelmetContainer";
import { RootState } from "../../redux/store";
import { SidebarProvider } from "../../HOCs/SidebarProvider";

function Admin(props) {
  const navigate = useNavigate();
  const role = useSelector(selectRole);
  const user = useSelector((state: RootState) => state.user);
  const [sidebarOpen, setSidebarOpen] = React.useState<boolean>(false);

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
      url.searchParams.append("userId", user?.userId || "");
      url.searchParams.append("plan", "admin");
      window.history.replaceState({}, "", url.toString());
    }

     
  }, [navigate, user?.userId, user?.user?.planType, role]);

  const initialSections: string[] = [
    "customers",
    "management",
    "configuration",
    "blog",
    "reports",
    "orders",
    "petitions",
    "donations",
  ];

  return (
    <SidebarProvider initialSections={initialSections}>
      <HelmetContainer title="Vbitez - Admin" />
      <div className="Admincontainer">
        <Navbar openSidebar={openSidebar} />
        <div className="InnerContainer">{props.children}</div>
        <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
      </div>
    </SidebarProvider>
  );
}

Admin.propTypes = {
  children: PropTypes.node.isRequired,
};

export default Admin;
