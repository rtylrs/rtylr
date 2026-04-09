import React from "react";
import { Outlet } from "react-router-dom";
import SettingsSidebar from "../SettingsSidebar";
import SettingsNavbar from "../SettingsNavbar";

const SettingsLayout: React.FC = () => {
  return (
    <div className="min-h-screen bg-admin-bg grid grid-rows-[auto_1fr] grid-cols-[220px_1fr]">
      <div className="col-span-2 row-start-1">
        <SettingsNavbar />
      </div>
      <div className="row-start-2 col-start-1 border-r border-black/8 bg-white">
        <SettingsSidebar />
      </div>
      <div className="row-start-2 col-start-2 p-8 overflow-auto">
        <Outlet />
      </div>
    </div>
  );
};

export default SettingsLayout;
