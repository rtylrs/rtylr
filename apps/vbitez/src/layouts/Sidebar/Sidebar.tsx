import { useLocation, useNavigate } from "react-router-dom";
import {
  SidebarHeader,
  SidebarUser,
  SidebarMenu,
  SidebarMenuItem,
} from "./components";
import { Puzzle } from "lucide-react";

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <div className={`vb-sidebar`}>
      <SidebarHeader />
      <SidebarMenu />
      <div style={{ padding: "8px 6px" }}>
        <SidebarMenuItem
          icon={Puzzle}
          text="Addons"
          isActive={location.pathname === "/app/business/addons"}
          onClick={() => navigate("/app/business/addons")}
        />
      </div>
      <SidebarUser />
    </div>
  );
};

export default Sidebar;