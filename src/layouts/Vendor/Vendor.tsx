import React from "react";
import Navbar from "./VendorNavbar";
import SideBar from "./VendorSidebar";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import HelmetContainer from "../../components/HelmetContainer";
import { selectRole } from "../../redux/reducers/userSlice";
import FreeAccount from "../../components/FreeAccount";
import { RootState } from "../../redux/store";
import { motion } from "framer-motion";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faWhatsapp } from "@fortawesome/free-brands-svg-icons";
import { SidebarProvider } from "../../HOCs/SidebarProvider";

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
      return;
    }

    const updateUrlParams = () => {
      const url = new URL(window.location.href);

      if (!url.searchParams.has("userId")) {
        url.searchParams.append("userId", (user?.userId || "").substring(0, 6));
      }

      if (!url.searchParams.has("plan")) {
        url.searchParams.append(
          "plan",
          role !== "admin" ? user?.user?.planType || "" : "admin"
        );
      }

      if (mode === "dev" && !url.searchParams.has("mode")) {
        url.searchParams.append("mode", "dev");
      }

      window.history.replaceState({}, "", url.toString());
    };

    updateUrlParams();
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
    "appearance",
    "pages",
  ];

  return (
    <SidebarProvider initialSections={initialSections}>
      <HelmetContainer title={`${user?.user?.name || "User"}'s Dashboard`} />
      {user?.user?.planType === "free" && <FreeAccount />}
      <div className="VendorContainer">
        <Navbar sidebarOpen={sidebarOpen} openSidebar={openSidebar} />
        <div className="InnerContainer-vendor">{props.children}</div>
        <SideBar sidebarOpen={sidebarOpen} closeSidebar={closeSidebar} />
      </div>
    </SidebarProvider>
  );
};

export default Vendor;
